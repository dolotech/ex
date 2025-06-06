// WsClient —— 火币公共 WS 客户端（自动重连、双向心跳、并发安全写）
// © 2025 Nicole Trading — fixed on 2025-05-05
package huobi

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	readTimeout   = 30 * time.Second // 必须 > 20 s
	writeTimeout  = 10 * time.Second
	keepAliveTick = 15 * time.Second // 小于 LB idle
)

/* ---------- 可选项 ---------- */

type SpotTickerMessage struct {
	Ch   string   `json:"ch"`
	Ts   int64    `json:"ts"`
	Tick SpotTick `json:"tick"`
}

type SpotTick struct {
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Amount    float64 `json:"amount"`
	Vol       float64 `json:"vol"`
	Count     int     `json:"count"`
	Bid       float64 `json:"bid"`
	BidSize   float64 `json:"bidSize"`
	Ask       float64 `json:"ask"`
	AskSize   float64 `json:"askSize"`
	LastPrice float64 `json:"lastPrice"`
	LastSize  float64 `json:"lastSize"`
}
type MarketMessage struct {
	Ch   string `json:"ch"`
	Ts   int64  `json:"ts"`
	Tick Tick   `json:"tick"`
}

type Tick struct {
	Id            int       `json:"id"`
	Mrid          int       `json:"mrid"`
	Open          float64   `json:"open"`
	Close         float64   `json:"close"`
	High          float64   `json:"high"`
	Low           float64   `json:"low"`
	Amount        float64   `json:"amount"`
	Vol           float64   `json:"vol"`
	TradeTurnover float64   `json:"trade_turnover"`
	Count         int       `json:"count"`
	Asks          []float64 `json:"asks"`
	Bids          []float64 `json:"bids"`
}

type Options struct {
	url           string
	subscriber    func(*websocket.Conn, ...string) error // 重连成功后重新订阅
	subscribeArgs []string
}

func SubscribeMarket(conn *websocket.Conn, symbols ...string) error {
	for _, symbol := range symbols {
		bs, err := json.Marshal(map[string]any{
			"sub": "market." + strings.Replace(symbol, "USDT", "-USDT", 1) + ".detail",
		})
		if err != nil {
			return err
		}
		if err = conn.WriteMessage(websocket.TextMessage, bs); err != nil {
			return err
		}
	}
	return nil
}
func SubscribeSpotMarket(conn *websocket.Conn, symbols ...string) error {
	for _, symbol := range symbols {
		bs, err := json.Marshal(map[string]any{
			"sub": "market." + strings.ToLower(symbol) + ".detail",
		})
		if err != nil {
			return err
		}
		if err = conn.WriteMessage(websocket.TextMessage, bs); err != nil {
			return err
		}
	}
	return nil
}
func SetURL(url string) func(*Options) { return func(o *Options) { o.url = url } }
func SetSubscriber(f func(*websocket.Conn, ...string) error, a []string) func(*Options) {
	return func(o *Options) { o.subscriber, o.subscribeArgs = f, a }
}
func SetSpot() func(o *Options) {
	return SetURL("wss://api.huobi.pro/ws")
}

/* ---------- 客户端结构 ---------- */

type WsClient struct {
	Options

	ctx    context.Context
	cancel context.CancelFunc

	dialer *websocket.Dialer
	conn   *websocket.Conn

	writeMu    sync.Mutex   // 写锁，保证并发安全
	reDialMu   sync.Mutex   // 防止同时触发多次重连
	alivePing  *time.Ticker // 客户端主动 ping 定时器
	handle     func([]byte) error
	pingCancel context.CancelFunc
	closeCh    chan struct{}
	isRedial   int32 // 0 = 空闲，1 = 正在重连
}

/* ---------- 构造函数 ---------- */

func NewWsClient(parent context.Context, handle func([]byte) error, opts ...func(*Options)) (*WsClient, error) {
	if parent == nil {
		parent = context.Background()
	}

	c := &WsClient{
		closeCh: make(chan struct{}, 1),
		dialer: &websocket.Dialer{
			ReadBufferSize:    4096 * 4,
			HandshakeTimeout:  10 * time.Second,
			EnableCompression: false, // 建议关闭，避免 1003 Unsupported Data
		},
		handle: handle,
	}
	c.ctx, c.cancel = context.WithCancel(parent)

	for _, o := range opts {
		o(&c.Options)
	}
	if c.url == "" {
		c.url = "wss://api.hbdm.com/linear-swap-ws"
	}

	// 首次拨号
	if err := c.reDial(); err != nil {
		return nil, err
	}

	return c, nil
}

/* ---------- 对外方法 ---------- */

// Close 线程安全关闭
func (c *WsClient) Close() error {
	select {
	case <-c.closeCh:
	default:
		close(c.closeCh) // ★ 通知所有子协程退出
		c.cancel()
		if c.pingCancel != nil {
			c.pingCancel()
		}
		if c.alivePing != nil {
			c.alivePing.Stop()
		}
		return c.conn.Close()
	}
	return nil
}

/* ============================== 读循环 ============================== */

func (c *WsClient) readLoop() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // readLoop 退出时终止所有 worker
	// 使用轻量级 worker pool，解耦 socket IO 与业务处理
	handleCh := make(chan []byte, 4096)
	for i := 0; i < 32; i++ {
		go func() {
			for {
				select {
				case bs := <-handleCh:
					_ = c.handle(bs)
				case <-c.closeCh:
					return
				case <-ctx.Done(): // 监听 ctx 取消信号
					return
				}

			}
		}()
	}

	for {
		_, bs, err := c.conn.ReadMessage()
		if err != nil {

			// 明确处理特定错误类型
			// if websocket.IsUnexpectedCloseError(err,
			// 	websocket.CloseAbnormalClosure,
			// 	websocket.CloseMessageTooBig) { // 处理消息过大
			// 	zap.S().Warnf("非预期关闭: %v", err)
			// }

			zap.S().Warn("读取失败:", err)
			time.Sleep(time.Second)
			c.triggerReDial() // 触发重连
			return            // ★ 立刻退出，避免再次读取同一个已失败的 conn
		}

		// 每收到任何数据都延长读超时，防止服务器误判断线
		if err := c.conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
			c.triggerReDial() // 触发重连
			return
		}

		// gunzip（火币所有业务数据均压缩）
		if unzipped, e := ParseGzip(bs); e == nil {
			bs = unzipped
		} else {
			zap.S().Warn("gzip 解压失败:", e)
			continue
		}

		// 处理服务器心跳 ping
		ping := struct {
			Ping int64 `json:"ping"`
		}{}
		if err = json.Unmarshal(bs, &ping); err == nil && ping.Ping > 0 {
			if err = c.writePong(); err != nil {
				zap.S().Error(err)
				c.triggerReDial() // 触发重连
				return
			}
			continue
		}

		// 保证 socket 读不被阻塞；若 backlog 满则丢弃最旧消息
		select {
		case <-c.closeCh:
			return
		case handleCh <- bs:
		default:
			zap.S().Warnf("handle backlog 已满，当前堆积: %d，丢弃一条消息", len(handleCh))
		}
	}
}

// 向服务器回 pong
func (c *WsClient) writePong() error {
	data, err := json.Marshal(map[string]any{
		"pong": time.Now().UnixMilli(),
	})
	if err != nil {
		return err
	}
	if err := c.safeWrite(websocket.TextMessage, data); err != nil {
		zap.S().Warn("发送 pong 失败:", err)
		return err
	}
	// 重置读超时

	return c.conn.SetReadDeadline(time.Now().Add(readTimeout))
}

/* ============================ 拨号与重连 ============================ */

// triggerReDial 在任何读/写失败时调用，异步重连
func (c *WsClient) triggerReDial() {
	if !atomic.CompareAndSwapInt32(&c.isRedial, 0, 1) { // 已在重连
		return
	}
	select {
	case <-c.closeCh:
		return
	default:
	}
	if c.conn != nil {
		_ = c.conn.Close() // 显式关闭旧连接
	}

	go func() {
		defer atomic.StoreInt32(&c.isRedial, 0)
		c.reDialMu.Lock()
		defer c.reDialMu.Unlock()

		if err := c.reDial(); err != nil {
			zap.S().Error("重连失败终止:", err)
			_ = c.Close()
			return
		}
	}()
}

// 带指数退避的重连流程
func (c *WsClient) reDial() error {
	backoff := time.Second
	for {
		if err := c.dial(); err != nil {
			if backoff < 32*time.Second {
				backoff += 2 * time.Second
			}
			zap.S().Warnf("拨号失败: %s，%s 后重试", err, backoff)
			select {
			case <-time.After(backoff):
				continue
			case <-c.ctx.Done():
				return c.ctx.Err()
			}
		}

		zap.S().Infof("重连成功后订阅 %v ", c.subscribeArgs)

		// 重订阅
		if c.subscriber != nil {
			if err := c.subscriber(c.conn, c.subscribeArgs...); err != nil {
				zap.S().Error("重订阅失败:", err)
				_ = c.conn.Close()
				continue
			}

			zap.S().Infof("订阅成功")
		}

		// ★ 重连成功后重新启动读循环
		go c.readLoop()

		return nil
	}
}

// dial 建立一次新的 WS 连接
func (c *WsClient) dial() error {
	conn, _, err := c.dialer.DialContext(c.ctx, c.url, nil)
	if err != nil {
		return err
	}
	// websocket.go 连接建立处（如 reDial 函数）添加
	const maxMessageSize = 4 * 1024 * 1024 // 16MB（根据业务调整）
	conn.SetReadLimit(maxMessageSize)      // 关键修复！
	// 读超时 70 秒
	_ = conn.SetReadDeadline(time.Now().Add(readTimeout))
	// 若对方主动 pong，也刷新超时
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(readTimeout))
		return nil
	})

	// 服务器发控制帧 Ping -> 立即回 Pong
	conn.SetPingHandler(func(appData string) error {
		_ = conn.SetReadDeadline(time.Now().Add(readTimeout))
		return c.safeWrite(websocket.PongMessage, []byte(appData))
	})
	c.startAlivePing() // 启动客户端保活
	// 替换旧连接
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.conn = conn
	return nil
}

/* -------------------- goroutine 安全写 -------------------- */

func (c *WsClient) safeWrite(t int, data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_ = c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err := c.conn.WriteMessage(t, data); err != nil {
		zap.S().Warn("写入失败:", err)
		// 立刻触发重连
		c.triggerReDial()
		return err
	}
	return nil
}

/* ---------- gzip 解压辅助 ---------- */
func ParseGzip(bs []byte) ([]byte, error) {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, bs)
	r, err := gzip.NewReader(b)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ---------------- 客户端保活 ----------------
func (c *WsClient) startAlivePing() {
	if c.pingCancel != nil {
		c.pingCancel()
	} // 结束旧协程
	var pingCtx context.Context
	pingCtx, c.pingCancel = context.WithCancel(c.ctx)

	t := time.NewTicker(keepAliveTick)
	go func() {
		defer t.Stop()
		for {
			select {
			case <-pingCtx.Done():
				return
			case <-t.C:
				_ = c.safeWrite(websocket.PingMessage, nil)
			}
		}
	}()
}
