// Package common：底层 WebSocket 连接维护
package common

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/339-Labs/v3-bitget-api-sdk-go/config"
	"github.com/339-Labs/v3-bitget-api-sdk-go/constants"
	"github.com/339-Labs/v3-bitget-api-sdk-go/internal"
	"github.com/339-Labs/v3-bitget-api-sdk-go/logging/applogger"
	"github.com/339-Labs/v3-bitget-api-sdk-go/model"
	"github.com/gorilla/websocket"
)

// ----------- 心跳 & 重连参数 -------------
const (
	pingInterval     = 25 * time.Second // 官方：≤30s 必须 ping
	pongTimeout      = 35 * time.Second // 10s 未收到 pong 判定掉线
	reconnectBackoff = 3 * time.Second  // 重连退避
	loginTimeout     = 5 * time.Second  // 登录等待
)

// BitgetBaseWsClient 仅供内部使用
type BitgetBaseWsClient struct {
	// 连接
	ws       *websocket.Conn
	lastPong time.Time
	mu       sync.RWMutex // 保护 ws

	// 订阅 / 回调
	scribeMap map[model.SubscribeReq]OnReceive // 专属回调
	allSub    *model.Set
	subMu     sync.RWMutex

	// 业务配置 / 状态
	cfg       *config.BitgetConfig
	needLogin bool
	loginOK   bool

	// 通用回调
	defListener       OnReceive
	errorListener     OnReceive
	reconnectListener OnReceive

	// 协程控制
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	reconnecting bool       // ★新增：避免并发重连
	recMu        sync.Mutex // ★新增
}

// Init：初始化
func (b *BitgetBaseWsClient) Init(cfg *config.BitgetConfig, needLogin bool, reconnectListener OnReceive) *BitgetBaseWsClient {
	b.cfg = cfg
	b.reconnectListener = reconnectListener
	b.needLogin = needLogin
	b.scribeMap = make(map[model.SubscribeReq]OnReceive)
	b.allSub = model.NewSet()
	b.lastPong = time.Now()
	b.ctx, b.cancel = context.WithCancel(context.Background())
	return b
}

// ---------------- 对上层暴露 ----------------

func (b *BitgetBaseWsClient) SetListener(l, errL OnReceive) {
	b.defListener = l
	b.errorListener = errL
}

// Send / SendByType
func (b *BitgetBaseWsClient) Send(msg string) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.ws == nil {
		return fmt.Errorf("no active ws")
	}
	b.ws.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return b.ws.WriteMessage(websocket.TextMessage, []byte(msg))
}
func (b *BitgetBaseWsClient) SendByType(req model.WsBaseReq) error {
	j, err := internal.ToJson(req)
	if err != nil {
		return err
	}
	return b.Send(j)
}

// 订阅管理
func (b *BitgetBaseWsClient) AddSub(r model.SubscribeReq, l OnReceive) {
	b.subMu.Lock()
	defer b.subMu.Unlock()
	if l != nil {
		b.scribeMap[r] = l
	}
	b.allSub.Add(r)
}
func (b *BitgetBaseWsClient) DelSub(r model.SubscribeReq) {
	b.subMu.Lock()
	defer b.subMu.Unlock()
	delete(b.scribeMap, r)
	b.allSub.Remove(r)
}

// ---------------- 连接 / 协程 ----------------

// ConnectWebSocket：建立连接 + 登录 + 续订
func (b *BitgetBaseWsClient) ConnectWebSocket() {
	conn, _, err := websocket.DefaultDialer.Dial(b.cfg.WsUrl, nil)
	if err != nil {
		applogger.Error("dial ws err: %v", err)
		return
	}
	conn.SetReadLimit(1 << 20)
	b.mu.Lock()
	b.ws = conn
	b.mu.Unlock()
	applogger.Info("ws connected")

	// 登录
	if b.needLogin {
		if err := b.login(); err != nil {
			applogger.Error("login fail: %v", err)
		}
	}
	// 续订
	b.resubscribe()
}

// StartReadLoop：读协程
func (b *BitgetBaseWsClient) StartReadLoop() {
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		for {
			select {
			case <-b.ctx.Done():
				return
			default:
			}
			b.mu.RLock()
			ws := b.ws
			b.mu.RUnlock()
			if ws == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			ws.SetReadDeadline(time.Now().Add(pongTimeout + 5*time.Second))
			_, data, err := ws.ReadMessage()
			if err != nil {
				applogger.Error("read err: %v, reconnecting", err)
				b.Reconnect()
				continue
			}
			b.handleMessage(string(data))
		}
	}()
}

// StartPingLoop：心跳协程
// StartPingLoop：心跳协程
func (b *BitgetBaseWsClient) StartPingLoop() {
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()

		tk := time.NewTicker(pingInterval)
		defer tk.Stop()

		// ★新增：局部变量，确保同一实例 5 秒内只打一次 “timeout” 日志
		var lastTimeoutLog time.Time

		for {
			select {
			case <-b.ctx.Done():
				return
			case <-tk.C:
				// 距离上次活动 > pongTimeout → 判超时
				if time.Since(b.lastPong) > pongTimeout {
					// ★同一实例若 5 s 内已记录过，不重复打印
					if time.Since(lastTimeoutLog) > 5*time.Second {
						applogger.Info("pong timeout, reconnecting")
						lastTimeoutLog = time.Now()
					}
					b.Reconnect()
					continue
				}
				// 发送 ping
				if err := b.Send("ping"); err != nil {
					applogger.Error("ping send err: %v", err)
					b.Reconnect()
				}
			}
		}
	}()
}

// Reconnect：统一重连入口（加锁防并发）
func (b *BitgetBaseWsClient) Reconnect() {
	b.recMu.Lock()
	if b.reconnecting { // ★若已在重连，无需重复
		b.recMu.Unlock()
		return
	}
	b.reconnecting = true
	b.recMu.Unlock()

	defer func() { // ★重连结束后解锁
		b.recMu.Lock()
		b.reconnecting = false
		b.recMu.Unlock()
	}()

	b.mu.Lock()
	if b.ws != nil {
		_ = b.ws.Close()
		b.ws = nil
	}
	b.mu.Unlock()

	time.Sleep(reconnectBackoff)
	b.ConnectWebSocket()
	if b.reconnectListener != nil {
		b.reconnectListener("reconnected")
	}
	// ★重置最近活跃时间，给下一轮心跳留足间隔
	b.lastPong = time.Now()
}

// Close：主程序退出时调用
func (b *BitgetBaseWsClient) Close() {
	b.cancel()
	b.mu.Lock()
	if b.ws != nil {
		_ = b.ws.Close()
	}
	b.mu.Unlock()
	b.wg.Wait()
}

// ---------------- 内部逻辑 ----------------

// login：鉴权
func (b *BitgetBaseWsClient) login() error {
	ts := internal.TimesStampSec()
	sign := new(Signer).Init(b.cfg.SecretKey).
		Sign(constants.WsAuthMethod, constants.WsAuthPath, "", ts)
	req := model.WsBaseReq{
		Op: constants.WsOpLogin,
		Args: []any{model.WsLoginReq{
			ApiKey:     b.cfg.ApiKey,
			Passphrase: b.cfg.PASSPHRASE,
			Timestamp:  ts,
			Sign:       sign,
		}},
	}
	if err := b.SendByType(req); err != nil {
		return err
	}
	// 等待 login event
	timer := time.NewTimer(loginTimeout)
	defer timer.Stop()
	for {
		if b.loginOK {
			return nil
		}
		select {
		case <-timer.C:
			return fmt.Errorf("login timeout")
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// resubscribe：重连后恢复所有订阅
func (b *BitgetBaseWsClient) resubscribe() {
	b.subMu.RLock()
	defer b.subMu.RUnlock()
	if b.allSub.Len() == 0 {
		return
	}
	var args []any
	for _, v := range b.allSub.List() {
		args = append(args, v.(model.SubscribeReq))
	}
	_ = b.SendByType(model.WsBaseReq{
		Op:   constants.WsOpSubscribe,
		Args: args,
	})
}

// handleMessage：分发消息
func (b *BitgetBaseWsClient) handleMessage(raw string) {
	if raw == "pong" { // 心跳回包
		b.lastPong = time.Now()
		return
	}
	msg := internal.JSONToMap(raw)

	// 登录成功
	if ev, ok := msg["event"]; ok && ev == "login" {
		b.loginOK = true
		return
	}

	// 错误码
	if code, ok := msg["code"]; ok && int(code.(float64)) != 0 {
		b.errorListener(raw)
		return
	}
	// 有 data 字段的行情推送
	if _, ok := msg["data"]; ok {
		b.dispatch(msg["arg"], raw)
		return
	}
	// 其余交给默认回调
	b.defListener(raw)
}

// dispatch：根据 arg 定位专属回调
func (b *BitgetBaseWsClient) dispatch(arg any, raw string) {
	m := arg.(map[string]any)
	req := model.SubscribeReq{
		InstType: fmt.Sprintf("%v", m["instType"]),
		Channel:  fmt.Sprintf("%v", m["channel"]),
		InstId:   fmt.Sprintf("%v", m["instId"]),
	}
	b.subMu.RLock()
	l, ok := b.scribeMap[req]
	b.subMu.RUnlock()
	if ok && l != nil {
		l(raw)
	} else {
		b.defListener(raw)
	}
}

// ---------------- 类型 ----------------

type OnReceive func(message string)
