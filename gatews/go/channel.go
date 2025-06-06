// subscribe.go
package gatews

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

/*------------------ 订阅相关结构 ------------------*/

type SubscribeOptions struct {
	ID          int64 `json:"id"`
	IsReConnect bool  `json:"-"`
}

/*------------------ 外部订阅 API ------------------*/

func (ws *WsService) Subscribe(channel string, payload []string) error {
	return ws.SubscribeWithOption(channel, payload, nil)
}

func (ws *WsService) SubscribeWithOption(channel string, payload []string, op *SubscribeOptions) error {
	// 鉴权频道但无 key/secret
	if (ws.conf.Key == "" || ws.conf.Secret == "") && authChannel[channel] {
		return newAuthEmptyErr()
	}

	// 初始化队列
	msgCh, ok := ws.msgChs.Load(channel)
	if !ok {
		msgCh = make(chan *UpdateMsg, 1)
		go ws.receiveCallMsg(channel, msgCh.(chan *UpdateMsg))
		ws.msgChs.Store(channel, msgCh)
	}

	return ws.newBaseChannel(channel, payload, msgCh.(chan *UpdateMsg), op)
}

/*------------------ 内部核心 ------------------*/

func (ws *WsService) newBaseChannel(channel string, payload []string, bch chan *UpdateMsg, op *SubscribeOptions) error {
	if err := ws.baseSubscribe(Subscribe, channel, payload, op); err != nil {
		return err
	}
	ws.once.Do(ws.readMsg) // 只启动一次读协程
	return nil
}

func (ws *WsService) baseSubscribe(event, channel string, payload []string, op *SubscribeOptions) error {
	ts := time.Now().Unix()

	// ★ 使用真实 event 计算签名
	hash := hmac.New(sha512.New, []byte(ws.conf.Secret))
	hash.Write([]byte(fmt.Sprintf("channel=%s&event=%s&time=%d", channel, event, ts)))

	req := Request{
		Time:    ts,
		Channel: channel,
		Event:   event,
		Payload: payload,
		Auth: Auth{
			Method: AuthMethodApiKey,
			Key:    ws.conf.Key,
			Secret: hex.EncodeToString(hash.Sum(nil)),
		},
	}
	if op != nil {
		req.Id = &op.ID
	}

	// 发送
	b, err := json.Marshal(req)
	if err != nil {
		zap.S().Errorf("序列化失败：%v", err)
		return err
	}
	if err = ws.safeWrite(websocket.TextMessage, b); err != nil {
		zap.S().Errorf("订阅发送失败：%v", err)
		return err
	}

	// ping / time 不记录历史
	if strings.HasSuffix(channel, ".ping") || strings.HasSuffix(channel, ".time") {
		return nil
	}

	// ★ 仅非重连场景写历史
	if op == nil || !op.IsReConnect {
		v, _ := ws.conf.subscribeMsg.LoadOrStore(channel, []requestHistory{})
		h := v.([]requestHistory)
		h = append(h, requestHistory{Channel: channel, Event: event, Payload: payload, op: op})
		ws.conf.subscribeMsg.Store(channel, h)
	}

	return nil
}

/*------------------ 读协程 ------------------*/

func (ws *WsService) readMsg() {
	go func() {
		defer func() {
			if e := recover(); e != nil {
				zap.S().Error(e, string(debug.Stack()))
			}
			_ = ws.Client.Close()
		}()

		ws.refreshDeadline() // 初始 watchdog

		for {
			select {
			case <-ws.Ctx.Done():
				zap.S().Info("读协程退出")
				return
			default:
				_, message, err := ws.Client.ReadMessage()
				if err != nil {
					if isNetErr(err) || err == io.ErrUnexpectedEOF {
						zap.S().Warnf("读错误：%v，重连中…", err)
						if e := ws.reconnect(); e != nil {
							zap.S().Errorf("重连失败：%v", e)
							return
						}
						continue
					}
					zap.S().Errorf("读取异常：%v", err)
					return
				}

				// 任意消息到来都刷新 ReadDeadline（★）
				ws.refreshDeadline()

				var raw UpdateMsg
				if err := json.Unmarshal(message, &raw); err != nil {
					zap.S().Warnf("反序列化失败：%v, 原始:%s", err, string(message))
					continue
				}
				if raw.Channel == "" {
					continue
				}

				// 分发
				if ch, ok := ws.msgChs.Load(raw.Channel); ok {
					select {
					case <-ws.Ctx.Done():
						return
					case ch.(chan *UpdateMsg) <- &raw:
					}
				}
			}
		}
	}()
}

/*------------------ 业务回调 ------------------*/

type callBack func(*UpdateMsg)

func (ws *WsService) SetCallBack(channel string, call callBack) {
	if call != nil {
		ws.calls.Store(channel, call)
	}
}

func (ws *WsService) receiveCallMsg(channel string, msgCh chan *UpdateMsg) {
	defer func() {
		if e := recover(); e != nil {
			zap.S().Error(e, string(debug.Stack()))
		}
	}()
	for {
		select {
		case <-ws.Ctx.Done():
			return
		case msg := <-msgCh:
			if cb, ok := ws.calls.Load(channel); ok {
				cb.(callBack)(msg)
			}
		}
	}
}

/*------------------ ReadDeadline 操作 ------------------*/

// refreshDeadline 刷新读超时
func (ws *WsService) refreshDeadline() {
	_ = ws.Client.SetReadDeadline(time.Now().Add(ws.conf.PingInterval * 4))
}
