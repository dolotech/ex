// Package bybit —— 通用 WebSocket 客户端（支持自动重连、心跳、并发安全）
// ★保持外部接口不变：NewWsClient() 即刻建连，Connect() 用于阻塞等待★
package bybit

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

/* =============== 连接可选项 =============== */

type Options struct {
	url           string
	subscriber    func(*websocket.Conn, ...string) error // 首次连上后的订阅动作
	subscribeArgs []string
}

func SetURL(url string) func(*Options) {
	return func(o *Options) { o.url = url }
}

func SetSubscriber(fn func(*websocket.Conn, ...string) error, args []string) func(*Options) {
	return func(o *Options) {
		o.subscriber = fn
		o.subscribeArgs = args
	}
}

/* =============== 客户端对象 =============== */

const handleQueueSize = 4096 // ★统一缓冲区大小★

type WsClient struct {
	Options
	ctx    context.Context
	cancel context.CancelFunc

	dialer    *websocket.Dialer
	conn      *websocket.Conn
	writeMu   sync.Mutex // ★写锁，防止并发 write 崩溃★
	connected int32      // 0=断开 1=连接

	handle      func([]byte) error // 业务处理回调
	handleQueue chan []byte        // ★内部统一分发队列★
}

/* =============== 构造 & 启动 =============== */

// NewWsClient：创建后即自动连接并启动所有协程
func NewWsClient(parent context.Context, handle func([]byte) error, opts ...func(*Options)) (*WsClient, error) {
	ctx, cancel := context.WithCancel(parent)

	c := &WsClient{
		ctx:         ctx,
		cancel:      cancel,
		dialer:      &websocket.Dialer{},
		handle:      handle,
		handleQueue: make(chan []byte, handleQueueSize),
	}

	for _, opt := range opts {
		opt(&c.Options)
	}
	if c.url == "" {
		c.url = "wss://stream.bytick.com/realtime_public"
	}

	// 首次拨号
	if err := c.redialBlocking(); err != nil {
		return nil, err
	}

	c.spawnLoops()
	return c, nil
}

/* =============== 对外控制 =============== */

// Connect：阻塞到退出，可供业务层等待
func (c *WsClient) Connect() error {
	<-c.ctx.Done()
	return c.ctx.Err()
}

// Close：安全关闭
func (c *WsClient) Close() error {
	c.cancel()
	if c.conn != nil {
		_ = c.conn.Close()
	}
	return nil
}

/* =============== 内部协程 =============== */

func (c *WsClient) spawnLoops() {
	/* ---- 读循环 ---- */
	go func() {
		for {
			_, bs, err := c.conn.ReadMessage()
			if err != nil {
				atomic.StoreInt32(&c.connected, 0)
				zap.S().Error("read:", err)
				if e := c.redialBlocking(); e != nil {
					zap.S().Error("redial fatal:", e)
					return
				}
				continue
			}
			c.dispatch(bs)
		}
	}()

	/* ---- 心跳循环 ---- */
	go func() {
		tk := time.NewTicker(30 * time.Second)
		defer tk.Stop()
		for {
			select {
			case <-c.ctx.Done():
				return
			case <-tk.C:
				if atomic.LoadInt32(&c.connected) == 1 {
					if err := c.safeWrite(websocket.TextMessage, []byte(`{"op":"ping"}`)); err != nil {
						zap.S().Error("ping:", err)
						if e := c.redialBlocking(); e != nil {
							zap.S().Error("redial fatal:", e)
							return
						}
					}
				}
			}
		}
	}()

	/* ---- 业务 worker 池 ---- */
	const worker = 16
	for i := 0; i < worker; i++ {
		go func() {
			for {
				select {
				case bs := <-c.handleQueue:
					if bs == nil {
						return
					}
					if err := c.handle(bs); err != nil {
						zap.S().Error("handle:", err)
					}
				case <-c.ctx.Done():
					return
				}
			}
		}()
	}
}

/* ---- 数据分发 ---- */

func (c *WsClient) dispatch(bs []byte) {
	select {
	case c.handleQueue <- bs:
	default: // 队列满丢弃，避免阻塞读循环
		zap.S().Warn("handleQueue full, drop msg")
	case <-c.ctx.Done():
	}
}

/* ---- 并发安全写 ---- */

func (c *WsClient) safeWrite(t int, data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteMessage(t, data)
}

/* ---- 拨号 + 自动重连 ---- */

func (c *WsClient) redialBlocking() error {
	backoff := time.Second
	for {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		default:
		}

		if err := c.dial(); err != nil {
			zap.S().Errorf("dial %s: %s", c.url, err)
			time.Sleep(backoff)
			if backoff < 30*time.Second {
				backoff *= 2
			}
			continue
		}

		if c.subscriber != nil {
			if err := c.subscriber(c.conn, c.subscribeArgs...); err != nil {
				zap.S().Error("subscribe:", err)
				time.Sleep(backoff)
				if backoff < 30*time.Second {
					backoff *= 2
				}
				continue
			}
		}
		return nil // 订阅成功
	}
}

func (c *WsClient) dial() error {
	conn, _, err := c.dialer.DialContext(c.ctx, c.url, nil)
	if err != nil {
		return err
	}
	atomic.StoreInt32(&c.connected, 1)

	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.conn = conn
	zap.S().Infof("connected: %s", conn.RemoteAddr())
	return nil
}

/* ---- 旧接口兼容性工具函数 ---- */

func sendOpArgs(conn *websocket.Conn, op, prefix string, symbols ...string) error {
	args := make([]string, len(symbols))
	for i, s := range symbols {
		args[i] = prefix + s
	}
	bs, _ := json.Marshal(map[string]interface{}{"op": op, "args": args})
	return conn.WriteMessage(websocket.TextMessage, bs)
}

func sendSpotReq(conn *websocket.Conn, topic, symbol string) error {
	req := map[string]interface{}{
		"topic": topic,
		"event": "sub",
		"params": map[string]any{
			"symbol": symbol,
			"binary": false,
		},
	}
	bs, _ := json.Marshal(req)
	return conn.WriteMessage(websocket.TextMessage, bs)
}
