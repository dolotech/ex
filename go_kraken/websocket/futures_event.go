package websocket

import (
	"encoding/json"
	"go.uber.org/zap"
)

func (k *Kraken) handleFuturesEvent(msg []byte) error {
	var event EventType
	if err := json.Unmarshal(msg, &event); err != nil {
		return err
	}

	switch event.Event {
	case FUTURES_EventPong:
		return k.handleEventPong(msg)
	case FUTURES_Alert:
		var message Message
		var ticker FuturesAlert
		if err := json.Unmarshal(msg, &ticker); err != nil {
			return err
		}
		message.ChannelName = event.Feed
		k.msg <- message.toUpdate(ticker)
		return nil
	case FUTURES_Info:
		var message Message
		var ticker FuturesInfo
		if err := json.Unmarshal(msg, &ticker); err != nil {
			return err
		}
		message.ChannelName = event.Feed
		k.msg <- message.toUpdate(ticker)
		return nil

	case FUTURES_Subscribed:
		var message Message
		var ticker FuturesSubscribed
		if err := json.Unmarshal(msg, &ticker); err != nil {
			return err
		}
		message.ChannelName = event.Feed
		k.msg <- message.toUpdate(ticker)
		return nil
	}

	switch event.Feed {
	case FUTURES_Ticker:
		var message Message
		var ticker FuturesTicker
		if err := json.Unmarshal(msg, &ticker); err != nil {
			return err
		}
		message.Pair = ticker.ProductId
		message.ChannelName = event.Feed
		k.msg <- message.toUpdate(ticker)

	case
		FUTURES_CANDLES_1M_SNAPSHOT,
		FUTURES_CANDLES_3M_SNAPSHOT,
		FUTURES_CANDLES_5M_SNAPSHOT,
		FUTURES_CANDLES_15M_SNAPSHOT,
		FUTURES_CANDLES_30M_SNAPSHOT,
		FUTURES_CANDLES_1H_SNAPSHOT,
		FUTURES_CANDLES_2H_SNAPSHOT,
		FUTURES_CANDLES_4H_SNAPSHOT,
		FUTURES_CANDLES_6H_SNAPSHOT,
		FUTURES_CANDLES_12H_SNAPSHOT,
		FUTURES_CANDLES_1D_SNAPSHOT:
		var message Message
		var candle FuturesCandle
		if err := json.Unmarshal(msg, &candle); err != nil {
			return err
		}
		message.ChannelName = event.Feed
		k.msg <- message.toUpdate(candle)
	case FUTURES_CANDLES_1M,
		FUTURES_CANDLES_3M,
		FUTURES_CANDLES_5M,
		FUTURES_CANDLES_15M,
		FUTURES_CANDLES_30M,
		FUTURES_CANDLES_1H,
		FUTURES_CANDLES_2H,
		FUTURES_CANDLES_4H,
		FUTURES_CANDLES_6H,
		FUTURES_CANDLES_12H,
		FUTURES_CANDLES_1D:
		var message Message
		var candle FuturesCandle
		if err := json.Unmarshal(msg, &candle); err != nil {
			return err
		}
		message.ChannelName = event.Feed
		k.msg <- message.toUpdate(candle)
	default:
		zap.S().Warnf("unknown event: %s", msg)
	}
	return nil
}
