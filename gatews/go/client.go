// ws_service.go
package gatews

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

/*------------------ 常量 ------------------*/

const (
	defaultMaxRetry        = 5
	defaultPingInterval    = 10 * time.Second
	defaultHandshakeTimout = 30 * time.Second // ★ 延长
	defaultNetTimeout      = 30 * time.Second // ★ 延长
	wsWriteWait            = 5 * time.Second  // 写超时
)

/*------------------ 核心结构 ------------------*/

type WsService struct {
	mu            sync.Mutex
	Ctx           context.Context
	cancel        context.CancelFunc
	Client        *websocket.Conn
	reconnectCall chan struct{}
	once          sync.Once
	msgChs        *sync.Map
	calls         *sync.Map
	conf          *ConnConf
}

// ConnConf 所有配置
type ConnConf struct {
	subscribeMsg     sync.Map
	URL              string
	Key              string
	Secret           string
	MaxRetryConn     int
	SkipTlsVerify    bool
	ShowReconnectMsg bool
	PingInterval     time.Duration
}

/*------------------ 构造函数 ------------------*/

func NewWsService(parent context.Context, conf *ConnConf, reconnect chan struct{}) (*WsService, error) {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)

	def := defaultConnConf()
	if conf != nil {
		conf = mergeConf(def, conf)
	} else {
		conf = def
	}

	conn, _, err := dial(conf)
	if err != nil {
		return nil, err
	}

	ws := &WsService{
		reconnectCall: reconnect,
		Ctx:           ctx,
		cancel:        cancel,
		Client:        conn,
		conf:          conf,
		calls:         new(sync.Map),
		msgChs:        new(sync.Map),
	}

	// watchdog & ping
	ws.setPongHandler()
	go ws.activePing()

	return ws, nil
}

func defaultConnConf() *ConnConf {
	return &ConnConf{
		URL:              BaseUrl,
		MaxRetryConn:     defaultMaxRetry,
		PingInterval:     defaultPingInterval,
		ShowReconnectMsg: true,
	}
}

func mergeConf(def, user *ConnConf) *ConnConf {
	if user.URL == "" {
		user.URL = def.URL
	}
	if user.MaxRetryConn == 0 {
		user.MaxRetryConn = def.MaxRetryConn
	}
	if user.PingInterval == 0 {
		user.PingInterval = def.PingInterval
	}
	return user
}

/*------------------ 连接 / 重连 ------------------*/

func dial(c *ConnConf) (*websocket.Conn, *http.Response, error) {
	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: defaultHandshakeTimout,
		NetDialContext: (&net.Dialer{
			Timeout:   defaultNetTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		EnableCompression: true,
	}
	if c.SkipTlsVerify {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return dialer.Dial(c.URL, nil)
}

func (ws *WsService) reconnect() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	retry, backoff := 0, time.Second
	_ = ws.Client.Close()

	for {
		if retry >= ws.conf.MaxRetryConn {
			return fmt.Errorf("reconnect exceeded max retry")
		}
		conn, _, err := dial(ws.conf)
		if err != nil {
			retry++
			zap.S().Warnf("重连失败(%d/%d)：%v，%v 后再试", retry, ws.conf.MaxRetryConn, err, backoff)
			time.Sleep(backoff)
			if backoff < 30*time.Second {
				backoff *= 2
			}
			continue
		}

		ws.Client = conn
		ws.setPongHandler()
		if ws.conf.ShowReconnectMsg {
			zap.S().Info("重连成功")
		}

		if ws.reconnectCall != nil {
			select {
			case ws.reconnectCall <- struct{}{}:
			default:
			}
		}

		ws.resubscribe()
		go ws.activePing()
		return nil
	}
}

/*------------------ watchdog & ping ------------------*/

// setPongHandler 设置 pong 处理并刷新读超时
func (ws *WsService) setPongHandler() {
	_ = ws.Client.SetReadDeadline(time.Now().Add(ws.conf.PingInterval * 4))
	ws.Client.SetPongHandler(func(string) error {
		_ = ws.Client.SetReadDeadline(time.Now().Add(ws.conf.PingInterval * 4))
		return nil
	})
}

// activePing 定时发送 **WebSocket ping + JSON ping**
func (ws *WsService) activePing() {
	tk := time.NewTicker(ws.conf.PingInterval)
	defer tk.Stop()

	for {
		select {
		case <-ws.Ctx.Done():
			return
		case <-tk.C:
			// 写控制帧 ping（★）
			_ = ws.Client.WriteControl(
				websocket.PingMessage,
				[]byte("heartbeat"),
				time.Now().Add(wsWriteWait),
			)

			// GateIO 业务层 ping
			now := time.Now().Unix()
			spot := fmt.Sprintf(`{"time":%d,"channel":"spot.ping"}`, now)
			future := fmt.Sprintf(`{"time":%d,"channel":"futures.ping"}`, now)

			if err := ws.safeWrite(websocket.TextMessage, []byte(spot)); err != nil {
				zap.S().Warnf("Ping 写入失败: %v", err)
				continue
			}
			_ = ws.safeWrite(websocket.TextMessage, []byte(future))
		}
	}
}

/*------------------ 写封装 ------------------*/

func (ws *WsService) safeWrite(t int, data []byte) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	_ = ws.Client.SetWriteDeadline(time.Now().Add(wsWriteWait))
	if err := ws.Client.WriteMessage(t, data); err != nil {
		if isNetErr(err) {
			_ = ws.reconnect()
		}
		return err
	}
	return nil
}

/*------------------ 订阅恢复 ------------------*/

func (ws *WsService) resubscribe() {
	ws.conf.subscribeMsg.Range(func(_, v any) bool {
		zap.S().Debugf("gateio恢复订阅 %+v", v)
		for _, r := range v.([]requestHistory) {
			if r.op == nil {
				r.op = &SubscribeOptions{}
			}
			r.op.IsReConnect = true
			if err := ws.baseSubscribe(r.Event, r.Channel, r.Payload, r.op); err != nil {
				zap.S().Warnf("恢复订阅 [%s] 失败：%v", r.Channel, err)
			}
		}
		return true
	})
}

/*------------------ 工具 ------------------*/

func isNetErr(err error) bool {
	if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
		return true
	}
	_, ok := err.(net.Error)
	return ok
}
