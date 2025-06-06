package okx

import (
	"net/http"
	"sync"
	"time"
)

type APIConfig struct {
	HttpClient    *http.Client
	Endpoint      string
	ApiKey        string
	ApiSecretKey  string
	ApiPassphrase string //for okex.com v3 api
	ClientId      string //for bitstamp.net , huobi.pro

	Lever float64 //杠杆倍数 , for future
}

const OKExV5WsPublicEndpoint = "wss://ws.okx.com:8443/ws/v5/public"

type OKExV5WsPublic struct {
	cfg *APIConfig
	*WsBuilder
	once   *sync.Once
	WsConn *WsConn
	hand   func([]byte) error
}

func NewOKExV5WsPublic(cfg *APIConfig, hand func([]byte) error, connected func(err error)) *OKExV5WsPublic {
	if cfg == nil {
		cfg = &APIConfig{Endpoint: OKExV5WsPublicEndpoint}
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = OKExV5WsPublicEndpoint
	}
	if cfg.HttpClient == nil {
		cfg.HttpClient = http.DefaultClient
	}

	pub := &OKExV5WsPublic{
		cfg:  cfg,
		once: new(sync.Once),
		hand: hand,
	}

	pub.WsBuilder = NewWsBuilder().
		WsUrl(cfg.Endpoint).
		ReconnectInterval(time.Second).
		AutoReconnect().
		Heartbeat(func() []byte { return []byte("ping") }, 15*time.Second).
		ConnectedHandleFunc(connected).  // 外部可感知“连通/重连”事件
		DecompressFunc(FlateDecompress). // OKX 使用 permessage-deflate
		ProtoHandleFunc(pub.handle)

	return pub
}

func (p *OKExV5WsPublic) ConnectWs() { p.once.Do(func() { p.WsConn = p.WsBuilder.Build() }) }

func (p *OKExV5WsPublic) handle(bs []byte) error {
	if string(bs) == "pong" { // 心跳
		return nil
	}
	if p.hand != nil {
		return p.hand(bs)
	}
	return nil
}

func (p *OKExV5WsPublic) Subscribe(req map[string]any) error {
	p.ConnectWs()
	return p.WsConn.Subscribe(req)
}
