// Package ws：业务层 WebSocket 客户端
package ws

import (
	"strings"

	"github.com/339-Labs/v3-bitget-api-sdk-go/common"
	"github.com/339-Labs/v3-bitget-api-sdk-go/config"
	"github.com/339-Labs/v3-bitget-api-sdk-go/constants"
	"github.com/339-Labs/v3-bitget-api-sdk-go/model"
)

// BitgetWsClient 对外暴露的顶层客户端
type BitgetWsClient struct {
	bitgetBaseWsClient *common.BitgetBaseWsClient
}

// Init：保持旧签名不变，内部改用增强版 BaseWsClient
func (c *BitgetWsClient) Init(
	cfg *config.BitgetConfig,
	needLogin bool,
	listener common.OnReceive,
	errorListener common.OnReceive,
	reconnectListener common.OnReceive,
) *BitgetWsClient {

	c.bitgetBaseWsClient = new(common.BitgetBaseWsClient).Init(cfg, needLogin, reconnectListener)
	c.bitgetBaseWsClient.SetListener(listener, errorListener)
	c.bitgetBaseWsClient.ConnectWebSocket() // 首次连接
	c.bitgetBaseWsClient.StartReadLoop()    // 读协程
	c.bitgetBaseWsClient.StartPingLoop()    // 心跳协程
	return c
}

// Connect：业务可显式调用以强制重连
func (c *BitgetWsClient) Connect() *BitgetWsClient {
	c.bitgetBaseWsClient.Reconnect()
	return c
}

// ---------------- 订阅 / 退订 ----------------

// Subscribe：带专属回调的订阅
func (c *BitgetWsClient) Subscribe(reqs []model.SubscribeReq, l common.OnReceive) {
	for _, r := range reqs {
		r = normalizeReq(r)
		c.bitgetBaseWsClient.AddSub(r, l)
	}
	c.bitgetBaseWsClient.SendByType(model.WsBaseReq{
		Op:   constants.WsOpSubscribe,
		Args: toAnySlice(reqs),
	})
}

// SubscribeDef：仅订阅，不设置专属回调
func (c *BitgetWsClient) SubscribeDef(reqs []model.SubscribeReq) {
	for _, r := range reqs {
		r = normalizeReq(r)
		c.bitgetBaseWsClient.AddSub(r, nil)
	}
	c.bitgetBaseWsClient.SendByType(model.WsBaseReq{
		Op:   constants.WsOpSubscribe,
		Args: toAnySlice(reqs),
	})
}

// UnSubscribe：退订
func (c *BitgetWsClient) UnSubscribe(reqs []model.SubscribeReq) {
	for _, r := range reqs {
		r = normalizeReq(r)
		c.bitgetBaseWsClient.DelSub(r)
	}
	c.bitgetBaseWsClient.SendByType(model.WsBaseReq{
		Op:   constants.WsOpUnsubscribe,
		Args: toAnySlice(reqs),
	})
}

// ---------------- 工具函数 ----------------

func (c *BitgetWsClient) SendMessage(msg string)              { c.bitgetBaseWsClient.Send(msg) }
func (c *BitgetWsClient) SendMessageByType(r model.WsBaseReq) { _ = c.bitgetBaseWsClient.SendByType(r) }

func normalizeReq(r model.SubscribeReq) model.SubscribeReq {
	r.InstType = strings.ToUpper(r.InstType)
	r.InstId = strings.ToUpper(r.InstId)
	r.Channel = strings.ToLower(r.Channel)
	if r.Coin == "" {
		r.Coin = strings.ToLower(r.InstId)
	}
	return r
}
func toAnySlice[T any](in []T) []any {
	out := make([]any, 0, len(in))
	for _, v := range in {
		out = append(out, v)
	}
	return out
}
