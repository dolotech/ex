package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/aopoltorzhicky/go_kraken/rest"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

// Kraken -
type Kraken struct {
	Options
	url   string
	token string
	conn  *websocket.Conn

	readTimeout      time.Duration
	heartbeatTimeout time.Duration

	msg  chan Update
	stop chan struct{}

	lock sync.RWMutex
	// ★新增：记录最近一次收到任何数据 / pong 的时间
	lastPong  atomic.Int64  // UnixNano
	reconnect chan struct{} // 长驻、缓冲 1
	wg        sync.WaitGroup
}

// NewKraken -
func NewKraken(url string, opts ...func(*Options)) *Kraken {
	kraken := Kraken{
		url:              url,
		readTimeout:      30 * time.Second,
		heartbeatTimeout: 15 * time.Second,
		// subscriptions:    make(map[int64]*SubscriptionStatus),
		msg:       make(chan Update, 1024),
		stop:      make(chan struct{}, 1),
		reconnect: make(chan struct{}, 1),
	}
	kraken.lastPong.Store(time.Now().UnixNano())
	for _, apply := range opts {
		apply(&kraken.Options)
	}
	return &kraken
}

// 工具函数
func (k *Kraken) triggerReconnect() {
	select {
	case k.reconnect <- struct{}{}: // 首次/空闲才能写入
	default: // 已有信号排队，丢弃避免死锁
	}
}

// Connect to the Kraken API, this should only be called once.
func (k *Kraken) Connect() error {
	var err error
	k.conn, err = k.dial()
	if err != nil {
		return err
	}

	if err := k.resubscribe(); err != nil {
		return err
	}

	go k.managerThread()

	return nil
}

func (k *Kraken) dial() (*websocket.Conn, error) {
	dialer := websocket.Dialer{
		Subprotocols:    []string{"p1", "p2"},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		Proxy:           http.ProxyFromEnvironment,
	}

	c, resp, err := dialer.Dial(k.url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	k.lastPong.Store(time.Now().UnixNano())
	c.SetPongHandler(func(_ string) error {
		k.lastPong.Store(time.Now().UnixNano())
		return nil
	})
	// Kraken 不要求客户端处理 Ping 帧，但保险起见回 Pong
	// c.SetPingHandler(func(appData string) error {
	// 	k.lock.RLock()
	// 	conn := k.conn
	// 	k.lock.RUnlock()
	// 	if conn != nil {
	// 		_ = conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(time.Second))
	// 	}
	// 	return nil
	// })

	return c, nil
}

// managerThread —— 用指数退避代替固定 5s
func (k *Kraken) managerThread() {
	heartbeat := time.NewTicker(k.heartbeatTimeout)
	defer heartbeat.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	k.wg.Add(1)
	go k.listenSocket(ctx)
	backoff := time.Second // ★新增：指数退避初始值

	for {
		select {
		case <-k.reconnect:
			cancel() // 旧 goroutine 退出
			time.Sleep(backoff)
			k.wg.Wait() // 等干净
			k.safeCloseConn()

			if backoff < 32*time.Second {
				backoff *= 2 // ★改动：指数增长
			}
			var err error
			k.conn, err = k.dial()
			if err != nil {
				zap.S().Error(err)
				time.Sleep(time.Second)
				k.triggerReconnect() // 继续重试
				continue
			}
			backoff = time.Second // ★成功后复位

			if err := k.resubscribe(); err != nil {
				zap.S().Error(err)
				k.triggerReconnect() // 继续重试
				continue
			}

			ctx, cancel = context.WithCancel(context.Background())
			k.wg.Add(1)
			go k.listenSocket(ctx)

		case <-k.stop:
			return

		case <-heartbeat.C:
			// ★改动：若超时未收到数据，直接触发重连
			if time.Since(time.Unix(0, k.lastPong.Load())) > k.heartbeatTimeout+k.readTimeout {
				zap.S().Warn("pong timeout, reconnecting...")
				k.triggerReconnect()
				continue
			}

			if err := k.send(PingRequest{Event: EventPing}); err != nil {
				zap.S().Error(err)
				k.triggerReconnect()
			}
		}
	}
}

func (c *Kraken) resubscribe() error {
	// 重订阅
	if c.conn != nil && len(c.Subscribers) > 0 {
		// --- automatic re-subscription ---
		for _, sub := range c.Subscribers {
			if err := sub(c.conn); err != nil {
				zap.S().Warnw("resubscribe failed", "err", err)
			}
		}
		zap.S().Infof("订阅成功")
	}

	return nil
}

// Listen provides an atomic interface for receiving API messages.
// When a websocket connection is terminated, the publisher channel will close.
func (k *Kraken) Listen() <-chan Update {
	return k.msg
}
func (k *Kraken) safeCloseConn() {
	k.lock.Lock()
	if k.conn != nil {
		_ = k.conn.Close()
		k.conn = nil
	}
	k.lock.Unlock()
}

// Close - provides an interface for a user initiated shutdown.
func (k *Kraken) Close() error {
	k.safeCloseConn()
	close(k.stop)
	close(k.msg)
	return nil
}

// ★改动：send 设置写超时，避免长时间阻塞
func (k *Kraken) send(msg any) error {
	k.lock.RLock()
	c := k.conn
	k.lock.RUnlock()
	if c == nil {
		return errors.New("ws not connected")
	}
	data, _ := json.Marshal(msg)
	_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return c.WriteMessage(websocket.TextMessage, data)
}

func (k *Kraken) listenSocket(ctx context.Context) {
	defer k.wg.Done()

	for {
		k.lock.RLock()
		c := k.conn // 每次循环取最新的引用
		k.lock.RUnlock()
		if c == nil {
			return // 说明已触发 shutdown
		}
		c.SetReadDeadline(time.Now().Add(k.readTimeout))
		_, msg, err := c.ReadMessage()
		if err != nil {
			zap.S().Error(err)
			k.triggerReconnect()
			return
		}

		// 更新最后活跃时间
		k.lastPong.Store(time.Now().UnixNano()) // 无论啥数据都视作活跃

		if err := k.handleMessage(msg); err != nil {
			zap.S().Error(err)
		}
	}
}

func (k *Kraken) handleMessage(data []byte) error {
	if len(data) == 0 {
		return errors.Errorf("Empty response: %s", string(data))
	}

	// Futures / 现货 分流处理
	if ProdBaseFuturesURL == k.url {
		return k.handleFuturesEvent(data)
	}

	switch data[0] {
	case '[':
		return k.handleChannel(data)
	case '{':
		return k.handleEvent(data)
	default:
		return errors.Errorf("Unexpected message: %s", string(data))
	}
}

// --- 以下业务功能函数保持不变，仅省略注释 ---
func (k *Kraken) SubscribeTicker(pairs []string) error {
	return k.send(SubscriptionRequest{
		Event: EventSubscribe,
		Pairs: pairs,
		Subscription: Subscription{
			Name: ChanTicker,
		},
	})
}
func (k *Kraken) SubscribeFuturesCandles(pairs []string, peroid string) error {
	return k.send(SubscribeFutures{
		Event:      EventSubscribe,
		ProductIds: pairs,
		Feed:       fmt.Sprintf("%s%s", FUTURES_CANDLES_, peroid),
	})
}
func (k *Kraken) SubscribeFuturesTicker(pairs []string) error {
	return k.send(SubscribeFutures{
		Event:      EventSubscribe,
		ProductIds: pairs,
		Feed:       FUTURES_Ticker,
	})
}
func (k *Kraken) SubscribeCandles(pairs []string, interval int64) error {
	return k.send(SubscriptionRequest{
		Event: EventSubscribe,
		Pairs: pairs,
		Subscription: Subscription{
			Name:     ChanCandles,
			Interval: interval,
		},
	})
}
func (k *Kraken) SubscribeTrades(pairs []string) error {
	return k.send(SubscriptionRequest{
		Event: EventSubscribe,
		Pairs: pairs,
		Subscription: Subscription{
			Name: ChanTrades,
		},
	})
}
func (k *Kraken) SubscribeSpread(pairs []string) error {
	return k.send(SubscriptionRequest{
		Event: EventSubscribe,
		Pairs: pairs,
		Subscription: Subscription{
			Name: ChanSpread,
		},
	})
}
func (k *Kraken) SubscribeBook(pairs []string, depth int64) error {
	return k.send(SubscriptionRequest{
		Event: EventSubscribe,
		Pairs: pairs,
		Subscription: Subscription{
			Name:  ChanBook,
			Depth: depth,
		},
	})
}
func (k *Kraken) Unsubscribe(channelType string, pairs []string) error {
	return k.send(UnsubscribeRequest{
		Event: EventUnsubscribe,
		Pairs: pairs,
		Subscription: Subscription{
			Name: channelType,
		},
	})
}
func (k *Kraken) UnsubscribeCandles(pairs []string, interval int64) error {
	return k.send(UnsubscribeRequest{
		Event: EventUnsubscribe,
		Pairs: pairs,
		Subscription: Subscription{
			Name:     ChanCandles,
			Interval: interval,
		},
	})
}
func (k *Kraken) UnsubscribeBook(pairs []string, depth int64) error {
	return k.send(UnsubscribeRequest{
		Event: EventUnsubscribe,
		Pairs: pairs,
		Subscription: Subscription{
			Name:  ChanBook,
			Depth: depth,
		},
	})
}
func (k *Kraken) Authenticate(key, secret string) error {
	data, err := rest.New(key, secret).GetWebSocketsToken()
	if err != nil {
		return err
	}
	k.token = data.Token
	return nil
}
func (k *Kraken) subscribeToPrivate(channelName string) error {
	return k.send(AuthSubscriptionRequest{
		Event: EventSubscribe,
		Subs: AuthDataRequest{
			Name:  channelName,
			Token: k.token,
		},
	})
}
func (k *Kraken) SubscribeOwnTrades() error  { return k.subscribeToPrivate(ChanOwnTrades) }
func (k *Kraken) SubscribeOpenOrders() error { return k.subscribeToPrivate(ChanOpenOrders) }
func (k *Kraken) AddOrder(req AddOrderRequest) error {
	req.Event = EventAddOrder
	req.Token = k.token
	return k.send(req)
}
func (k *Kraken) CancelOrder(orderIDs []string) error {
	return k.send(CancelOrderRequest{
		AuthRequest: AuthRequest{
			Token: k.token,
			Event: EventCancelOrder,
		},
		TxID: orderIDs,
	})
}
func (k *Kraken) CancelAll() error {
	return k.send(AuthRequest{
		Token: k.token,
		Event: EventCancelAll,
	})
}
func (k *Kraken) CancelAllOrdersAfter(timeout int64) error {
	return k.send(CancelAllOrdersAfterRequest{
		AuthRequest: AuthRequest{
			Token: k.token,
			Event: EventCancelAllOrdersAfter,
		},
		Timeout: timeout,
	})
}
func (k *Kraken) EditOrder(req EditOrderRequest) error {
	req.Event = EventEditOrder
	req.Token = k.token
	return k.send(req)
}
