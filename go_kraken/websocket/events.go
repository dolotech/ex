package websocket

import (
	"encoding/json"

	"go.uber.org/zap"
)

func (k *Kraken) handleEvent(msg []byte) error {
	var event EventType
	if err := json.Unmarshal(msg, &event); err != nil {
		return err
	}

	switch event.Event {
	case EventPong:
		return k.handleEventPong(msg)
	case EventSystemStatus:
		return k.handleEventSystemStatus(msg)
	case EventSubscriptionStatus:
		return k.handleEventSubscriptionStatus(msg)
	case EventCancelOrderStatus:
		return k.handleEventCancelOrderStatus(msg)
	case EventAddOrderStatus:
		return k.handleEventAddOrderStatus(msg)
	case EventCancelAllStatus:
		return k.handleEventCancellAllStatus(msg)
	case EventCancelAllOrdersAfter:
		return k.handleEventCancellAllOrdersAfter(msg)
	case EventEditOrderStatus:
		return k.handleEventEditOrderStatus(msg)
	case EventHeartbeat:
	default:
		zap.S().Warnf("unknown event: %s", msg)
	}
	return nil
}

func (k *Kraken) handleEventPong(data []byte) error {
	var pong PongResponse
	return json.Unmarshal(data, &pong)
}

func (k *Kraken) handleEventSystemStatus(data []byte) error {
	var systemStatus SystemStatus
	if err := json.Unmarshal(data, &systemStatus); err != nil {
		return err
	}
	zap.S().Infof("Status: %s", systemStatus.Status)
	zap.S().Infof("Connection ID: %s", systemStatus.ConnectionID.String())
	zap.S().Infof("Version: %s", systemStatus.Version)
	return nil
}

func (k *Kraken) handleEventSubscriptionStatus(data []byte) error {
	var status SubscriptionStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return err
	}

	if status.Status == SubscriptionStatusError {
		zap.S().Errorf("%s: %s", status.Error, status.Pair)
	} else {
		// zap.S().Infof("\tStatus: %s", status.Status)
		// zap.S().Infof("\tPair: %s", status.Pair)
		// zap.S().Infof("\tSubscription: %s", status.Subscription.Name)
		// zap.S().Infof("\tChannel ID: %d", status.ChannelID)
		// zap.S().Infof("\tReq ID: %s", status.ReqID)

		// if status.Status == SubscriptionStatusSubscribed {
		// 	k.subscriptions[status.ChannelID] = &status
		// } else if status.Status == SubscriptionStatusUnsubscribed {
		// 	delete(k.subscriptions, status.ChannelID)
		// }
	}
	return nil
}

func (k *Kraken) handleEventCancelOrderStatus(data []byte) error {
	var cancelOrderResponse CancelOrderResponse
	if err := json.Unmarshal(data, &cancelOrderResponse); err != nil {
		return err
	}

	switch cancelOrderResponse.Status {
	case StatusError:
		zap.S().Errorf(cancelOrderResponse.ErrorMessage)
	case StatusOK:
		zap.S().Debug(" Order successfully cancelled")
		k.msg <- Update{
			ChannelName: EventCancelOrder,
			Data:        cancelOrderResponse,
		}
	default:
		zap.S().Errorf("Unknown status: %s", cancelOrderResponse.Status)
	}
	return nil
}

func (k *Kraken) handleEventAddOrderStatus(data []byte) error {
	var addOrderResponse AddOrderResponse
	if err := json.Unmarshal(data, &addOrderResponse); err != nil {
		return err
	}

	switch addOrderResponse.Status {
	case StatusError:
		zap.S().Errorf(addOrderResponse.ErrorMessage)
	case StatusOK:
		zap.S().Debug("Order successfully sent")
		k.msg <- Update{
			ChannelName: EventAddOrder,
			Data:        addOrderResponse,
		}
	default:
		zap.S().Errorf("Unknown status: %s", addOrderResponse.Status)
	}
	return nil
}

func (k *Kraken) handleEventCancellAllStatus(data []byte) error {
	var cancelAllResponse CancelAllResponse
	if err := json.Unmarshal(data, &cancelAllResponse); err != nil {
		return err
	}

	switch cancelAllResponse.Status {
	case StatusError:
		zap.S().Errorf(cancelAllResponse.ErrorMessage)
	case StatusOK:
		zap.S().Debugf("%d orders cancelled", cancelAllResponse.Count)
		k.msg <- Update{
			ChannelName: EventCancelAllStatus,
			Data:        cancelAllResponse,
		}
	default:
		zap.S().Errorf("Unknown status: %s", cancelAllResponse.Status)
	}
	return nil
}

func (k *Kraken) handleEventCancellAllOrdersAfter(data []byte) error {
	var cancelAllResponse CancelAllOrdersAfterResponse
	if err := json.Unmarshal(data, &cancelAllResponse); err != nil {
		return err
	}

	switch cancelAllResponse.Status {
	case StatusError:
		zap.S().Errorf(cancelAllResponse.ErrorMessage)
	case StatusOK:
		k.msg <- Update{
			ChannelName: EventCancelAllOrdersAfter,
			Data:        cancelAllResponse,
		}
	default:
		zap.S().Errorf("Unknown status: %s", cancelAllResponse.Status)
	}
	return nil
}

func (k *Kraken) handleEventEditOrderStatus(data []byte) error {
	var editOrderResponse EditOrderResponse
	if err := json.Unmarshal(data, &editOrderResponse); err != nil {
		return err
	}

	switch editOrderResponse.Status {
	case StatusError:
		zap.S().Errorf(editOrderResponse.ErrorMessage)
	case StatusOK:
		zap.S().Debug("Order successfully edited")
		k.msg <- Update{
			ChannelName: EventEditOrder,
			Data:        editOrderResponse,
		}
	default:
		zap.S().Errorf("Unknown status: %s", editOrderResponse.Status)
	}
	return nil
}
