package websocket

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// One callback = one (re-)subscription step
type Subscriber func(*websocket.Conn) error
type Options struct {
	Url         string
	Subscribers []Subscriber // 重连成功后重新订阅
}

// Spot-candles helper
func WithSpotCandles(interval int64, pairs ...string) func(*Options) {
	return func(o *Options) {
		// capture params in the closure:
		o.Subscribers = append(o.Subscribers, func(conn *websocket.Conn) error {
			return SubscribeCandles(conn, interval, pairs...)
		})
	}
}

// 现货
func SubscribeCandles(conn *websocket.Conn, interval int64, pairs ...string) error {
	bs, err := json.Marshal(SubscriptionRequest{
		Event: EventSubscribe,
		Pairs: pairs,
		Subscription: Subscription{
			Name:     ChanCandles,
			Interval: interval,
		},
	})
	if err != nil {
		return err
	}
	if err = conn.WriteMessage(websocket.TextMessage, bs); err != nil {
		zap.S().Errorf("订阅失败:%v %v", string(bs), err)
		return err
	}
	return nil
}

// Futures-candles helper
func WithFuturesCandles(period string, pairs ...string) func(*Options) {
	return func(o *Options) {
		o.Subscribers = append(o.Subscribers, func(conn *websocket.Conn) error {
			return SubscribeFuturesCandles(conn, period, pairs...)
		})
	}
}

// 期货
func SubscribeFuturesCandles(conn *websocket.Conn, peroid string, pairs ...string) error {
	bs, err := json.Marshal(SubscribeFutures{
		Event:      EventSubscribe,
		ProductIds: pairs,
		Feed:       fmt.Sprintf("%s%s", FUTURES_CANDLES_, peroid),
	})
	if err != nil {
		return err
	}
	if err = conn.WriteMessage(websocket.TextMessage, bs); err != nil {
		zap.S().Errorf("订阅失败:%v %v", string(bs), err)
		return err
	}
	return nil
}

// Trades helper
func WithTrades(pairs ...string) func(*Options) {
	return func(o *Options) {
		o.Subscribers = append(o.Subscribers, func(conn *websocket.Conn) error {
			return SubscribeTrades(conn, pairs...)
		})
	}
}

func SubscribeTrades(conn *websocket.Conn, pairs ...string) error {
	bs, err := json.Marshal(SubscriptionRequest{
		Event: EventSubscribe,
		Pairs: pairs,
		Subscription: Subscription{
			Name: ChanTrades,
		},
	})
	if err != nil {
		return err
	}
	if err = conn.WriteMessage(websocket.TextMessage, bs); err != nil {
		zap.S().Errorf("订阅失败:%v %v", string(bs), err)
		return err
	}
	return nil
}

// Ticker helper
func WithFuturesTicker(pairs ...string) func(*Options) {
	return func(o *Options) {
		o.Subscribers = append(o.Subscribers, func(conn *websocket.Conn) error {
			return SubscribeFuturesTicker(conn, pairs...)
		})
	}
}

func SubscribeFuturesTicker(conn *websocket.Conn, pairs ...string) error {
	bs, err := json.Marshal(SubscribeFutures{
		Event:      EventSubscribe,
		ProductIds: pairs,
		Feed:       FUTURES_Ticker,
	})
	if err != nil {
		return err
	}
	if err = conn.WriteMessage(websocket.TextMessage, bs); err != nil {
		zap.S().Errorf("订阅失败:%v %v", string(bs), err)
		return err
	}
	return nil
}
