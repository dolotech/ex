package okx

import (
	"encoding/json"
	"errors"
	"math/rand/v2"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WsConfig struct {
	WsUrl                          string
	ProxyUrl                       string
	ReqHeaders                     map[string][]string //连接的时候加入的头部信息
	HeartbeatIntervalTime          time.Duration       //
	HeartbeatData                  func() []byte       //心跳数据2
	IsAutoReconnect                bool
	ProtoHandleFunc                func([]byte) error           //协议处理函数
	DecompressFunc                 func([]byte) ([]byte, error) //解压函数
	ErrorHandleFunc                func(err error)
	ConnectSuccessAfterSendMessage func() []byte //for reconnect
	ConnectedHandleFunc            func(err error)
	IsDump                         bool
	DisableEnableCompression       bool
	readDeadLineTime               time.Duration
	reconnectInterval              time.Duration
}

// NewDialer 返回一个默认配置的独立 Dialer，避免并发竞争。
func NewDialer() *websocket.Dialer {
	return &websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  8 * time.Second,
		EnableCompression: true,
		ReadBufferSize:    32 * 1024,
		WriteBufferSize:   32 * 1024,
	}
}

type WsConn struct {
	c *websocket.Conn
	WsConfig
	writeBufferChan        chan []byte
	pingMessageBufferChan  chan []byte
	pongMessageBufferChan  chan []byte
	closeMessageBufferChan chan []byte
	subs                   [][]byte
	close                  chan struct{}
	reConnectLock          *sync.Mutex
	writeLock              sync.Mutex // ★新增：全局写锁★
}

type WsBuilder struct {
	wsConfig *WsConfig
}

func NewWsBuilder() *WsBuilder {
	return &WsBuilder{&WsConfig{
		ReqHeaders:        make(map[string][]string, 1),
		reconnectInterval: time.Millisecond * 100,
	}}
}

func (b *WsBuilder) WsUrl(wsUrl string) *WsBuilder {
	b.wsConfig.WsUrl = wsUrl
	return b
}

func (b *WsBuilder) ProxyUrl(proxyUrl string) *WsBuilder {
	b.wsConfig.ProxyUrl = proxyUrl
	return b
}

func (b *WsBuilder) ReqHeader(key, value string) *WsBuilder {
	b.wsConfig.ReqHeaders[key] = append(b.wsConfig.ReqHeaders[key], value)
	return b
}

func (b *WsBuilder) AutoReconnect() *WsBuilder {
	b.wsConfig.IsAutoReconnect = true
	return b
}

func (b *WsBuilder) Dump() *WsBuilder {
	b.wsConfig.IsDump = true
	return b
}

func (b *WsBuilder) Heartbeat(heartbeat func() []byte, t time.Duration) *WsBuilder {
	b.wsConfig.HeartbeatIntervalTime = t
	b.wsConfig.HeartbeatData = heartbeat
	return b
}

func (b *WsBuilder) ReconnectInterval(t time.Duration) *WsBuilder {
	b.wsConfig.reconnectInterval = t
	return b
}

func (b *WsBuilder) ProtoHandleFunc(f func([]byte) error) *WsBuilder {
	b.wsConfig.ProtoHandleFunc = f
	return b
}

func (b *WsBuilder) DisableEnableCompression() *WsBuilder {
	b.wsConfig.DisableEnableCompression = true
	return b
}

func (b *WsBuilder) DecompressFunc(f func([]byte) ([]byte, error)) *WsBuilder {
	b.wsConfig.DecompressFunc = f
	return b
}

func (b *WsBuilder) ErrorHandleFunc(f func(err error)) *WsBuilder {
	b.wsConfig.ErrorHandleFunc = f
	return b
}

func (b *WsBuilder) ConnectedHandleFunc(msg func(err error)) *WsBuilder {
	b.wsConfig.ConnectedHandleFunc = msg
	return b
}

func (b *WsBuilder) ConnectSuccessAfterSendMessage(msg func() []byte) *WsBuilder {
	b.wsConfig.ConnectSuccessAfterSendMessage = msg
	return b
}

func (b *WsBuilder) Build() *WsConn {
	wsConn := &WsConn{WsConfig: *b.wsConfig}
	return wsConn.NewWs()
}
func (ws *WsConn) NewWs() *WsConn {
	/* ---- 1. 通道 & 锁：必须先初始化，后续任何路径都可安全写入 ---- */

	ws.close = make(chan struct{})
	ws.writeBufferChan = make(chan []byte, 16)
	ws.pingMessageBufferChan = make(chan []byte, 4)
	ws.pongMessageBufferChan = make(chan []byte, 4)
	ws.closeMessageBufferChan = make(chan []byte, 2)
	ws.reConnectLock = new(sync.Mutex)

	/* ---- 2. 读超时 ---- */
	if ws.HeartbeatIntervalTime == 0 {
		ws.readDeadLineTime = time.Second * 60
	} else {
		ws.readDeadLineTime = ws.HeartbeatIntervalTime * 2
	}

	/* ---- 3. 拨号 —— 带抖动的指数退避 ---- */
	const (
		maxRetry      = 10
		initBackoff   = 100 * time.Millisecond
		maxBackoff    = 2 * time.Second
		jitterPercent = 0.15
	)

	backoff := initBackoff
	for i := 1; i <= maxRetry; i++ {
		if err := ws.connect(); err == nil {
			zap.S().Infof("[ws][%s] dial OK", ws.WsUrl)
			goto READY
		}

		jitter := 1 + (rand.Float64()*2-1)*jitterPercent
		sleep := time.Duration(float64(backoff) * jitter)
		if sleep > maxBackoff {
			sleep = maxBackoff
		}
		zap.S().Warnf("[ws][%s] dial retry %d/%d in %v", ws.WsUrl, i, maxRetry, sleep)
		time.Sleep(sleep)

		if backoff < maxBackoff {
			backoff <<= 1
		}
	}

	zap.S().Errorf("[ws][%s] dial failed after %d retries", ws.WsUrl, maxRetry)
	return nil

READY:
	/* ---- 4. 协程启动 ---- */
	go ws.writeRequest()
	go ws.receiveMessage()

	/* ---- 5. 首次连通后自动发送订阅或自定义报文 ---- */
	if ws.ConnectSuccessAfterSendMessage != nil {
		msg := ws.ConnectSuccessAfterSendMessage()
		ws.SendMessage(msg)
		zap.S().Debugf("[ws][%s] sent ConnectSuccessAfter message=%s", ws.WsUrl, string(msg))
	}

	return ws
}

func (ws *WsConn) connect() error {
	dialer := NewDialer()
	if ws.ProxyUrl != "" {
		proxy, err := url.Parse(ws.ProxyUrl)
		if err == nil {
			zap.S().Errorf("[ws][%s] proxy url:%s", ws.WsUrl, proxy)
			dialer.Proxy = http.ProxyURL(proxy)
		} else {
			zap.S().Errorf("[ws][%s]parse proxy url [%s] err %s  ", ws.WsUrl, ws.ProxyUrl, err.Error())
		}
	}

	if ws.DisableEnableCompression {
		dialer.EnableCompression = false
	}

	wsConn, resp, err := dialer.Dial(ws.WsUrl, http.Header(ws.ReqHeaders))

	if ws.ConnectedHandleFunc != nil {
		ws.ConnectedHandleFunc(err)
	}
	if err != nil {
		zap.S().Errorf("[ws][%s] %s", ws.WsUrl, err.Error())
		if ws.IsDump && resp != nil {
			dumpData, _ := httputil.DumpResponse(resp, true)
			zap.S().Errorf("[ws][%s] %s", ws.WsUrl, string(dumpData))
		}
		return err
	}

	wsConn.SetReadDeadline(time.Now().Add(ws.readDeadLineTime))

	if ws.IsDump {
		dumpData, _ := httputil.DumpResponse(resp, true)
		zap.S().Errorf("[ws][%s] %s", ws.WsUrl, string(dumpData))
	}
	ws.c = wsConn
	return nil
}

func (ws *WsConn) reconnect() {
	select {
	case <-ws.close:
		return
	default:
	}
	ws.reConnectLock.Lock()
	defer ws.reConnectLock.Unlock()

	_ = ws.c.Close()
	backoff := time.Second
	for {
		if err := ws.connect(); err != nil {
			zap.S().Errorf("[ws][%s] reconnect fail: %s", ws.WsUrl, err)
			time.Sleep(backoff)
			if backoff < 32*time.Second {
				backoff += 2
			} else {
				zap.S().Errorf("[ws] [%s] retry connect 100 count fail , begin exiting. ", ws.WsUrl)
				ws.CloseWs()
				if ws.ErrorHandleFunc != nil {
					ws.ErrorHandleFunc(errors.New("retry reconnect fail"))
				}
			}
			continue
		}
		break
	}

	if ws.ConnectSuccessAfterSendMessage != nil {
		msg := ws.ConnectSuccessAfterSendMessage()
		ws.SendMessage(msg)
		zap.S().Errorf("[ws] [%s] execute the connect success after send message=%s", ws.WsUrl, string(msg))
		time.Sleep(time.Second) //wait response
	}

	for _, sub := range ws.subs {
		zap.S().Infof("[ws] re subscribe: %s", string(sub))
		ws.SendMessage(sub)
	}
	// }
}

func (ws *WsConn) writeRequest() {
	var (
		heartTimer *time.Timer
		// err        error
	)

	if ws.HeartbeatIntervalTime == 0 {
		heartTimer = time.NewTimer(time.Second * 8)
	} else {
		heartTimer = time.NewTimer(ws.HeartbeatIntervalTime)
	}

	for {
		select {
		case <-ws.close:
			return
		case d := <-ws.writeBufferChan:
			ws.safeWrite(websocket.TextMessage, d)
		case d := <-ws.pingMessageBufferChan:
			ws.safeWrite(websocket.PingMessage, d)
		case d := <-ws.pongMessageBufferChan:
			ws.safeWrite(websocket.PongMessage, d)
		case d := <-ws.closeMessageBufferChan:
			ws.safeWrite(websocket.CloseMessage, d)
		case <-heartTimer.C:
			if ws.HeartbeatIntervalTime > 0 {
				ws.safeWrite(websocket.TextMessage, ws.HeartbeatData())
				heartTimer.Reset(ws.HeartbeatIntervalTime)
			}
		}
	}
}
func (ws *WsConn) safeWrite(mt int, data []byte) {
	ws.writeLock.Lock()
	err := ws.c.WriteMessage(mt, data)
	ws.writeLock.Unlock()

	if err != nil {
		zap.S().Errorf("[ws][%s] write err: %s", ws.WsUrl, err)
		// FIX(OKX): 立即标记失联，交给 readLoop->reconnect
		_ = ws.c.Close()
	}
}

func (ws *WsConn) Subscribe(subEvent interface{}) error {
	data, err := json.Marshal(subEvent)
	if err != nil {
		zap.S().Errorf("[ws][%s] json encode error , %s", ws.WsUrl, err)
		return err
	}
	ws.writeBufferChan <- data
	ws.subs = append(ws.subs, data)
	return nil
}

func (ws *WsConn) SendMessage(msg []byte) {
	ws.writeBufferChan <- msg
}

func (ws *WsConn) SendPingMessage(msg []byte) {
	ws.pingMessageBufferChan <- msg
}

func (ws *WsConn) SendPongMessage(msg []byte) {
	ws.pongMessageBufferChan <- msg
}

func (ws *WsConn) SendCloseMessage(msg []byte) {
	ws.closeMessageBufferChan <- msg
}

func (ws *WsConn) SendJsonMessage(m any) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	ws.writeBufferChan <- data
	return nil
}

func (ws *WsConn) receiveMessage() {
	//exit
	ws.c.SetCloseHandler(func(code int, text string) error {
		zap.S().Errorf("[ws][%s] websocket exiting [code=%d , text=%s]", ws.WsUrl, code, text)
		//ws.CloseWs()
		return nil
	})

	ws.c.SetPongHandler(func(pong string) error {
		zap.S().Errorf("[%s] received [pong] %s", ws.WsUrl, pong)
		ws.c.SetReadDeadline(time.Now().Add(ws.readDeadLineTime))
		return nil
	})

	ws.c.SetPingHandler(func(ping string) error {
		zap.S().Errorf("[%s] received [ping] %s", ws.WsUrl, ping)
		ws.SendPongMessage([]byte(ping))
		ws.c.SetReadDeadline(time.Now().Add(ws.readDeadLineTime))
		return nil
	})

	for {
		select {
		case <-ws.close:
			zap.S().Errorf("[ws][%s] close websocket , exiting receive message goroutine.", ws.WsUrl)
			return
		default:
			t, msg, err := ws.c.ReadMessage()
			if err != nil {
				zap.S().Errorf("[ws][%s] %s", ws.WsUrl, err.Error())
				if ws.IsAutoReconnect {
					zap.S().Errorf("[ws][%s] Unexpected Closed , Begin Retry Connect.", ws.WsUrl)
					ws.reconnect()
					continue
				}

				if ws.ErrorHandleFunc != nil {
					ws.ErrorHandleFunc(err)
				}

				return
			}
			//			zap.S().Debug(string(msg))
			ws.c.SetReadDeadline(time.Now().Add(ws.readDeadLineTime))
			switch t {
			case websocket.TextMessage:
				ws.ProtoHandleFunc(msg)
			case websocket.BinaryMessage:
				if ws.DecompressFunc == nil {
					ws.ProtoHandleFunc(msg)
				} else {
					msg2, err := ws.DecompressFunc(msg)
					if err != nil {
						zap.S().Errorf("[ws][%s] decompress error %s", ws.WsUrl, err.Error())
					} else {
						ws.ProtoHandleFunc(msg2)
					}
				}
				//	case websocket.CloseMessage:
				//	ws.CloseWs()
			default:
				zap.S().Errorf("[ws][%s] error websocket message type , content is :\n %s \n", ws.WsUrl, string(msg))
			}
		}
	}
}

func (ws *WsConn) CloseWs() {
	select {
	case <-ws.close:
	default:
		ws.close <- struct{}{}
		close(ws.close)
		close(ws.writeBufferChan)
		close(ws.closeMessageBufferChan)
		close(ws.pingMessageBufferChan)
		close(ws.pongMessageBufferChan)

		err := ws.c.Close()
		if err != nil {
			zap.S().Errorf("[ws][", ws.WsUrl, "] close websocket error ,", err)
		}
	}
}
