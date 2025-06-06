package bybit

import (
	sjson "encoding/json"
	"strings"
	"time"
)

type BaseResult struct {
	RetCode         int         `json:"ret_code"`
	RetMsg          string      `json:"ret_msg"`
	ExtCode         string      `json:"ext_code"`
	Result          interface{} `json:"result"`
	TimeNow         string      `json:"time_now"`
	RateLimitStatus int         `json:"rate_limit_status"`
}

type Item struct {
	Price float64 `json:"price,string"`
	Size  float64 `json:"size"`
}

type OrderBook struct {
	Asks []Item    `json:"asks"`
	Bids []Item    `json:"bids"`
	Time time.Time `json:"time"`
}

type RawItem struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price,string"`
	Size   float64 `json:"size"`
	Side   string  `json:"side"` // Buy/Sell
}

type GetOrderBookResult struct {
	RetCode int       `json:"ret_code"`
	RetMsg  string    `json:"ret_msg"`
	ExtCode string    `json:"ext_code"`
	ExtInfo string    `json:"ext_info"`
	Result  []RawItem `json:"result"`
	TimeNow string    `json:"time_now"`
}

type OHLC struct {
	Symbol   string  `json:"symbol"`
	Interval string  `json:"interval"`
	OpenTime int64   `json:"open_time"`
	Open     float64 `json:"open,string"`
	High     float64 `json:"high,string"`
	Low      float64 `json:"low,string"`
	Close    float64 `json:"close,string"`
	Volume   float64 `json:"volume,string"`
	Turnover float64 `json:"turnover,string"`
}

type GetKlineResult struct {
	RetCode int    `json:"ret_code"`
	RetMsg  string `json:"ret_msg"`
	ExtCode string `json:"ext_code"`
	ExtInfo string `json:"ext_info"`
	Result  []OHLC `json:"result"`
	TimeNow string `json:"time_now"`
}

type Ticker struct {
	Symbol               string    `json:"symbol"`
	BidPrice             float64   `json:"bid_price,string"`
	AskPrice             float64   `json:"ask_price,string"`
	LastPrice            float64   `json:"last_price,string"`
	LastTickDirection    string    `json:"last_tick_direction"`
	PrevPrice24H         float64   `json:"prev_price_24h,string"`
	Price24HPcnt         float64   `json:"price_24h_pcnt,string"`
	HighPrice24H         float64   `json:"high_price_24h,string"`
	LowPrice24H          float64   `json:"low_price_24h,string"`
	PrevPrice1H          float64   `json:"prev_price_1h,string"`
	Price1HPcnt          float64   `json:"price_1h_pcnt,string"`
	MarkPrice            float64   `json:"mark_price,string"`
	IndexPrice           float64   `json:"index_price,string"`
	OpenInterest         int       `json:"open_interest"`
	OpenValue            float64   `json:"open_value,string"`
	TotalTurnover        float64   `json:"total_turnover,string"`
	Turnover24H          float64   `json:"turnover_24h,string"`
	TotalVolume          int64     `json:"total_volume"`
	Volume24H            int64     `json:"volume_24h"`
	FundingRate          float64   `json:"funding_rate,string"`
	PredictedFundingRate float64   `json:"predicted_funding_rate,string"`
	NextFundingTime      time.Time `json:"next_funding_time"`
	CountdownHour        int       `json:"countdown_hour"`
}

type GetTickersResult struct {
	RetCode int      `json:"ret_code"`
	RetMsg  string   `json:"ret_msg"`
	ExtCode string   `json:"ext_code"`
	ExtInfo string   `json:"ext_info"`
	Result  []Ticker `json:"result"`
	TimeNow string   `json:"time_now"`
}

type TradingRecord struct {
	ID     int       `json:"id"`
	Symbol string    `json:"symbol"`
	Price  float64   `json:"price"`
	Qty    int       `json:"qty"`
	Side   string    `json:"side"`
	Time   time.Time `json:"time"`
}

type GetTradingRecordsResult struct {
	RetCode int             `json:"ret_code"`
	RetMsg  string          `json:"ret_msg"`
	ExtCode string          `json:"ext_code"`
	ExtInfo string          `json:"ext_info"`
	Result  []TradingRecord `json:"result"`
	TimeNow string          `json:"time_now"`
}

type LeverageFilter struct {
	MinLeverage  int     `json:"min_leverage"`
	MaxLeverage  int     `json:"max_leverage"`
	LeverageStep float64 `json:"leverage_step,string"`
}

type PriceFilter struct {
	MinPrice float64 `json:"min_price,string"`
	MaxPrice float64 `json:"max_price,string"`
	TickSize float64 `json:"tick_size,string"`
}

type LotSizeFilter struct {
	MaxTradingQty float64 `json:"max_trading_qty"`
	MinTradingQty float64 `json:"min_trading_qty"`
	QtyStep       float64 `json:"qty_step"`
}

type SymbolInfo struct {
	Name           string         `json:"name"`
	BaseCurrency   string         `json:"base_currency"`
	QuoteCurrency  string         `json:"quote_currency"`
	PriceScale     int            `json:"price_scale"`
	TakerFee       float64        `json:"taker_fee,string"`
	MakerFee       float64        `json:"maker_fee,string"`
	LeverageFilter LeverageFilter `json:"leverage_filter"`
	PriceFilter    PriceFilter    `json:"price_filter"`
	LotSizeFilter  LotSizeFilter  `json:"lot_size_filter"`
}

type GetSymbolsResult struct {
	RetCode int          `json:"ret_code"`
	RetMsg  string       `json:"ret_msg"`
	ExtCode string       `json:"ext_code"`
	ExtInfo string       `json:"ext_info"`
	Result  []SymbolInfo `json:"result"`
	TimeNow string       `json:"time_now"`
}

type Balance struct {
	Equity           float64 `json:"equity"`
	AvailableBalance float64 `json:"available_balance"`
	UsedMargin       float64 `json:"used_margin"`
	OrderMargin      float64 `json:"order_margin"`
	PositionMargin   float64 `json:"position_margin"`
	OccClosingFee    float64 `json:"occ_closing_fee"`
	OccFundingFee    float64 `json:"occ_funding_fee"`
	WalletBalance    float64 `json:"wallet_balance"`
	RealisedPnl      float64 `json:"realised_pnl"`
	UnrealisedPnl    float64 `json:"unrealised_pnl"`
	CumRealisedPnl   float64 `json:"cum_realised_pnl"`
	GivenCash        float64 `json:"given_cash"`
	ServiceCash      float64 `json:"service_cash"`
}

type GetBalanceResult struct {
	RetCode          int                  `json:"ret_code"`
	RetMsg           string               `json:"ret_msg"`
	ExtCode          string               `json:"ext_code"`
	ExtInfo          string               `json:"ext_info"`
	Result           GetBalanceResultData `json:"result"`
	TimeNow          string               `json:"time_now"`
	RateLimitStatus  int                  `json:"rate_limit_status"`
	RateLimitResetMs int64                `json:"rate_limit_reset_ms"`
	RateLimit        int                  `json:"rate_limit"`
}

type GetBalanceResultData struct {
	BTC  Balance `json:"BTC"`
	ETH  Balance `json:"ETH"`
	EOS  Balance `json:"EOS"`
	XRP  Balance `json:"XRP"`
	USDT Balance `json:"USDT"`
}

type CreateOrderResult struct {
	RetCode         int    `json:"ret_code"`
	RetMsg          string `json:"ret_msg"`
	ExtCode         string `json:"ext_code"`
	Result          Order  `json:"result"`
	TimeNow         string `json:"time_now"`
	RateLimitStatus int    `json:"rate_limit_status"`
}

type OrderLite struct {
	OrderID string `json:"order_id"`
}

type ReplaceOrderResult struct {
	RetCode         int       `json:"ret_code"`
	RetMsg          string    `json:"ret_msg"`
	ExtCode         string    `json:"ext_code"`
	Result          OrderLite `json:"result"`
	TimeNow         string    `json:"time_now"`
	RateLimitStatus int       `json:"rate_limit_status"`
}

type CancelOrderResult struct {
	RetCode         int    `json:"ret_code"`
	RetMsg          string `json:"ret_msg"`
	ExtCode         string `json:"ext_code"`
	Result          Order  `json:"result"`
	TimeNow         string `json:"time_now"`
	RateLimitStatus int    `json:"rate_limit_status"`
}

type OrderListResultData struct {
	Data        []Order `json:"data"`
	CurrentPage int     `json:"current_page"`
	LastPage    int     `json:"last_page"`
}

type OrderListResult struct {
	RetCode         int                 `json:"ret_code"`
	RetMsg          string              `json:"ret_msg"`
	ExtCode         string              `json:"ext_code"`
	Result          OrderListResultData `json:"result"`
	TimeNow         string              `json:"time_now"`
	RateLimitStatus int                 `json:"rate_limit_status"`
}

// Order ...
type Order struct {
	OrderID     string  `json:"order_id"`
	StopOrderID string  `json:"stop_order_id"`
	UserID      int     `json:"user_id"`
	Symbol      string  `json:"symbol"`
	Side        string  `json:"side"`
	OrderType   string  `json:"order_type"`
	Price       float64 `json:"price"`
	Qty         float64 `json:"qty"`
	TimeInForce string  `json:"time_in_force"`
	//StopOrderType   string       `json:"stop_order_type,omitempty"`
	//StopPx          sjson.Number `json:"stop_px,omitempty"`
	OrderStatus string `json:"order_status"`
	//StopOrderStatus string       `json:"stop_order_status"`
	LastExecPrice float64 `json:"last_exec_price"`
	LeavesQty     float64 `json:"leaves_qty"`

	CumExecQty   float64 `json:"cum_exec_qty"`
	CumExecValue float64 `json:"cum_exec_value"`
	CumExecFee   float64 `json:"cum_exec_fee"`

	ReduceOnly     bool      `json:"reduce_only"`
	CloseOnTrigger bool      `json:"close_on_trigger"`
	OrderLinkID    string    `json:"order_link_id"`
	CreatedTime    time.Time `json:"create_time"`
	UpdatedTime    time.Time `json:"updated_time"`
	TakeProfit     float64   `json:"take_profit"`
	StopLoss       float64   `json:"stop_loss"`
	TpTriggerBy    string    `json:"tp_trigger_by"`
	SlTriggerBy    string    `json:"sl_trigger_by"`
}

type ExtFields struct {
	ReduceOnly  bool   `json:"reduce_only"`
	OpFrom      string `json:"op_from"`
	Remark      string `json:"remark"`
	OReqNum     int64  `json:"o_req_num"`
	XreqType    string `json:"xreq_type"`
	CrossStatus string `json:"cross_status,omitempty"`
}

type InExtFields struct {
	ReduceOnly  bool   `json:"reduce_only"`
	OpFrom      string `json:"op_from"`
	Remark      string `json:"remark"`
	OReqNum     int64  `json:"o_req_num"`
	XreqType    string `json:"xreq_type"`
	CrossStatus string `json:"cross_status,omitempty"`
}

func (e *ExtFields) MarshalJSON() ([]byte, error) {
	return json.Marshal(e)
}

func (e *ExtFields) UnmarshalJSON(b []byte) error {
	s := string(b)
	if strings.HasPrefix(s, "[") {
		return nil
	}
	o := InExtFields{}
	if err := json.Unmarshal(b, &o); err == nil {
		e.ReduceOnly = o.ReduceOnly
		e.OpFrom = o.OpFrom
		e.Remark = o.Remark
		e.OReqNum = o.OReqNum
		e.XreqType = o.XreqType
		e.CrossStatus = o.CrossStatus
		return nil
	} else {
		return err
	}
}

type GetLeverageResult struct {
	RetCode         int                     `json:"ret_code"`
	RetMsg          string                  `json:"ret_msg"`
	ExtCode         string                  `json:"ext_code"`
	Result          map[string]LeverageItem `json:"result"`
	TimeNow         string                  `json:"time_now"`
	RateLimitStatus int                     `json:"rate_limit_status"`
}

type LeverageItem struct {
	Leverage int `json:"leverage"`
}

type Position struct {
	UserID              uint    `json:"user_id"`              // 用戶ID
	Symbol              string  `json:"symbol"`               // 合約類型
	Side                string  `json:"side"`                 // 方向
	Size                float64 `json:"size"`                 // 倉位數量
	PositionValue       float64 `json:"position_value"`       // 當前倉位價值
	EntryPrice          float64 `json:"entry_price"`          // 平均開倉價
	LiqPrice            float64 `json:"liq_price"`            // 強平價格
	BustPrice           float64 `json:"bust_price"`           // 破產價格
	Leverage            float64 `json:"leverage"`             // 逐倉模式下, 值為用戶設置的杠桿；全倉模式下，值為當前風險限額下最大杠桿
	AutoAddMargin       uint8   `json:"auto_add_margin"`      // 是否 自動追加保證金
	IsIsolated          bool    `json:"is_isolated"`          // 是否逐倉，true-逐倉 false-全倉
	PositionMargin      float64 `json:"position_margin"`      // 倉位保證金
	OccClosingFee       float64 `json:"occ_closing_fee"`      // 預占用平倉手續費
	RealisedPnl         float64 `json:"realised_pnl"`         // 當日已結盈虧
	CumRealisedPnl      float64 `json:"cumRealised_pnl"`      // 累計已結盈虧
	FreeQty             float64 `json:"free_qty"`             // 可平倉位數量
	TpSlMode            string  `json:"tp_sl_mode"`           // 止盈止損模式
	DeleverageIndicator uint8   `json:"deleverage_indicator"` // 風險指示燈等級（1，2，3，4，5）
	UnrealisedPnl       float64 `json:"unrealised_pnl"`       // 未結盈虧
	RiskID              uint8   `json:"risk_id"`              // 風險限額ID
	TakeProfit          float64 `json:"take_profit"`          // 止盈價格
	StopLoss            float64 `json:"stop_loss"`            // 止損價格
	TrailingStop        float64 `json:"trailing_stop"`        // 追蹤止損（與當前價格的距離）
	PositionIdx         uint    `json:"position_idx"`         // Position idx, 用於在不同倉位模式下標識倉位：
	Mode                string  `json:"mode"`                 // 倉位模式: MergedSingle - 單倉模式 BothSide - 雙倉模式
}

type PositionExtFields struct {
	Remark string `json:"_remark"`
}

type PositionListResult struct {
	BaseResult
	ExtInfo interface{} `json:"ext_info"`
	Result  []Position  `json:"result"`
}

type GetPositionResult struct {
	BaseResult
	ExtInfo interface{} `json:"ext_info"`
	Result  []Position  `json:"result"`
}

type SetPositionModeResult struct {
	BaseResult
	ExtInfo interface{} `json:"ext_info"`
}

type OrderV2 struct {
	UserID        int          `json:"user_id"`
	OrderID       string       `json:"order_id"`
	Symbol        string       `json:"symbol"`
	Side          string       `json:"side"`
	OrderType     string       `json:"order_type"`
	Price         sjson.Number `json:"price"`
	Qty           float64      `json:"qty"`
	TimeInForce   string       `json:"time_in_force"`
	OrderStatus   string       `json:"order_status"`
	LastExecTime  sjson.Number `json:"last_exec_time"`
	LastExecPrice sjson.Number `json:"last_exec_price"`
	LeavesQty     float64      `json:"leaves_qty"`
	CumExecQty    float64      `json:"cum_exec_qty"`
	CumExecValue  sjson.Number `json:"cum_exec_value"`
	CumExecFee    sjson.Number `json:"cum_exec_fee"`
	RejectReason  string       `json:"reject_reason"`
	OrderLinkID   string       `json:"order_link_id"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

type CreateOrderV2Result struct {
	RetCode          int     `json:"ret_code"`
	RetMsg           string  `json:"ret_msg"`
	ExtCode          string  `json:"ext_code"`
	ExtInfo          string  `json:"ext_info"`
	Result           OrderV2 `json:"result"`
	TimeNow          string  `json:"time_now"`
	RateLimitStatus  int     `json:"rate_limit_status"`
	RateLimitResetMs int64   `json:"rate_limit_reset_ms"`
	RateLimit        int     `json:"rate_limit"`
}

type CancelOrderV2Result struct {
	RetCode          int     `json:"ret_code"`
	RetMsg           string  `json:"ret_msg"`
	ExtCode          string  `json:"ext_code"`
	ExtInfo          string  `json:"ext_info"`
	Result           OrderV2 `json:"result"`
	TimeNow          string  `json:"time_now"`
	RateLimitStatus  int     `json:"rate_limit_status"`
	RateLimitResetMs int64   `json:"rate_limit_reset_ms"`
	RateLimit        int     `json:"rate_limit"`
}

type CancelAllOrderV2Result struct {
	RetCode          int       `json:"ret_code"`
	RetMsg           string    `json:"ret_msg"`
	ExtCode          string    `json:"ext_code"`
	ExtInfo          string    `json:"ext_info"`
	Result           []OrderV2 `json:"result"`
	TimeNow          string    `json:"time_now"`
	RateLimitStatus  int       `json:"rate_limit_status"`
	RateLimitResetMs int64     `json:"rate_limit_reset_ms"`
	RateLimit        int       `json:"rate_limit"`
}

type QueryOrderResult struct {
	RetCode          int     `json:"ret_code"`
	RetMsg           string  `json:"ret_msg"`
	ExtCode          string  `json:"ext_code"`
	ExtInfo          string  `json:"ext_info"`
	Result           OrderV2 `json:"result"`
	TimeNow          string  `json:"time_now"`
	RateLimitStatus  int     `json:"rate_limit_status"`
	RateLimitResetMs int64   `json:"rate_limit_reset_ms"`
	RateLimit        int     `json:"rate_limit"`
}

type StopOrderV2 struct {
	ClOrdID           string       `json:"clOrdID"`
	UserID            int64        `json:"user_id"`
	Symbol            string       `json:"symbol"`
	Side              string       `json:"side"`
	OrderType         string       `json:"order_type"`
	Price             sjson.Number `json:"price"`
	Qty               float64      `json:"qty"`
	TimeInForce       string       `json:"time_in_force"`
	CreateType        string       `json:"create_type"`
	CancelType        string       `json:"cancel_type"`
	OrderStatus       string       `json:"order_status"`
	LeavesQty         float64      `json:"leaves_qty"`
	LeavesValue       string       `json:"leaves_value"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
	CrossStatus       string       `json:"cross_status"`
	CrossSeq          float64      `json:"cross_seq"`
	StopOrderType     string       `json:"stop_order_type"`
	TriggerBy         string       `json:"trigger_by"`
	BasePrice         sjson.Number `json:"base_price"`
	ExpectedDirection string       `json:"expected_direction"`
}

type CancelStopOrdersV2Result struct {
	RetCode          int           `json:"ret_code"`
	RetMsg           string        `json:"ret_msg"`
	ExtCode          string        `json:"ext_code"`
	ExtInfo          string        `json:"ext_info"`
	Result           []StopOrderV2 `json:"result"`
	TimeNow          string        `json:"time_now"`
	RateLimitStatus  int           `json:"rate_limit_status"`
	RateLimitResetMs int64         `json:"rate_limit_reset_ms"`
	RateLimit        int           `json:"rate_limit"`
}

type StopOrder struct {
	UserID          int64     `json:"user_id"`
	StopOrderStatus string    `json:"stop_order_status"`
	Symbol          string    `json:"symbol"`
	Side            string    `json:"side"`
	OrderType       string    `json:"order_type"`
	Price           float64   `json:"price"`
	Qty             float64   `json:"qty"`
	TimeInForce     string    `json:"time_in_force"`
	StopOrderType   string    `json:"stop_order_type"`
	TriggerBy       string    `json:"trigger_by"`
	BasePrice       float64   `json:"base_price"`
	OrderLinkID     string    `json:"order_link_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	StopPx          float64   `json:"stop_px"`
	StopOrderID     string    `json:"stop_order_id"`
}

type GetStopOrdersResultData struct {
	CurrentPage int         `json:"current_page"`
	LastPage    int         `json:"last_page"`
	Data        []StopOrder `json:"data"`
}

type GetStopOrdersResult struct {
	RetCode          int                     `json:"ret_code"`
	RetMsg           string                  `json:"ret_msg"`
	ExtCode          string                  `json:"ext_code"`
	Result           GetStopOrdersResultData `json:"result"`
	ExtInfo          interface{}             `json:"ext_info"`
	TimeNow          string                  `json:"time_now"`
	RateLimitStatus  int                     `json:"rate_limit_status"`
	RateLimitResetMs int64                   `json:"rate_limit_reset_ms"`
	RateLimit        int                     `json:"rate_limit"`
}

type InstrumentInfo struct {
	Topic       string             `json:"topic"`
	Type        string             `json:"type"`
	Data        InstrumentInfoData `json:"data"`
	CrossSeq    string             `json:"cross_seq"`
	TimestampE6 string             `json:"timestamp_e6"`
}

type InstrumentInfoData struct {
	Id                     int       `json:"id"`
	Symbol                 string    `json:"symbol"`
	LastPriceE4            string    `json:"last_price_e4"`
	LastPrice              string    `json:"last_price"`
	Bid1PriceE4            string    `json:"bid1_price_e4"`
	Bid1Price              string    `json:"bid1_price"`
	Ask1PriceE4            string    `json:"ask1_price_e4"`
	Ask1Price              string    `json:"ask1_price"`
	LastTickDirection      string    `json:"last_tick_direction"`
	PrevPrice24HE4         string    `json:"prev_price_24h_e4"`
	PrevPrice24H           float64   `json:"prev_price_24h,string"`
	Price24HPcntE6         string    `json:"price_24h_pcnt_e6"`
	HighPrice24HE4         string    `json:"high_price_24h_e4"`
	HighPrice24H           string    `json:"high_price_24h"`
	LowPrice24HE4          string    `json:"low_price_24h_e4"`
	LowPrice24H            string    `json:"low_price_24h"`
	PrevPrice1HE4          string    `json:"prev_price_1h_e4"`
	PrevPrice1H            string    `json:"prev_price_1h"`
	Price1HPcntE6          string    `json:"price_1h_pcnt_e6"`
	MarkPriceE4            string    `json:"mark_price_e4"`
	MarkPrice              string    `json:"mark_price"`
	IndexPriceE4           string    `json:"index_price_e4"`
	IndexPrice             string    `json:"index_price"`
	OpenInterestE8         string    `json:"open_interest_e8"`
	TotalTurnoverE8        string    `json:"total_turnover_e8"`
	Turnover24HE8          string    `json:"turnover_24h_e8"`
	TotalVolumeE8          string    `json:"total_volume_e8"`
	Volume24HE8            string    `json:"volume_24h_e8"`
	FundingRateE6          string    `json:"funding_rate_e6"`
	PredictedFundingRateE6 string    `json:"predicted_funding_rate_e6"`
	CrossSeq               string    `json:"cross_seq"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
	NextFundingTime        time.Time `json:"next_funding_time"`
	CountDownHour          string    `json:"count_down_hour"`
	FundingRateInterval    string    `json:"funding_rate_interval"`
	SettleTimeE9           string    `json:"settle_time_e9"`
	DelistingStatus        string    `json:"delisting_status"`
	Update                 Update    `json:"update"`
}

type Update []struct {
	Id              int       `json:"id"`
	Symbol          string    `json:"symbol"`
	LastPriceE4     string    `json:"last_price_e4"`
	LastPrice       float64   `json:"last_price,string"`
	Price24HPcntE6  float64   `json:"price_24h_pcnt_e6,string"`
	Price1HPcntE6   string    `json:"price_1h_pcnt_e6"`
	IndexPriceE4    string    `json:"index_price_e4"`
	IndexPrice      float64   `json:"index_price,string"`
	TotalTurnoverE8 string    `json:"total_turnover_e8"`
	Turnover24HE8   string    `json:"turnover_24h_e8"`
	TotalVolumeE8   string    `json:"total_volume_e8"`
	Volume24HE8     string    `json:"volume_24h_e8"`
	PrevPrice24H    float64   `json:"prev_price_24h,string"`
	CrossSeq        string    `json:"cross_seq"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

/*
symbol	string	幣對
binary	string	是否壓縮 true壓縮, false未壓縮（默認）
symbolName	string	幣對
t	number	時間（撮合引擎撮合时间）
s	string	幣對
c	string	收盤價
h	string	最高價
l	string	最低價
o	string	開盤價
v	string	成交量
qv	string	成交金額
m	string	漲幅
*/
type RealTimes struct {
	Topic  string `json:"topic"`
	Params struct {
		Symbol string `json:"symbol"`
		// Binary     bool   `json:"binary"`
		SymbolName string `json:"symbolName"`
	} `json:"params"`
	Data struct {
		T  int64  `json:"t"`
		S  string `json:"s"`
		O  string `json:"o"`
		H  string `json:"h"`
		L  string `json:"l"`
		C  string `json:"c"`
		V  string `json:"v"`
		Qv string `json:"qv"`
		M  string `json:"m"`
	} `json:"data"`
}

type Trade struct {
	Topic  string `json:"topic"`
	Params struct {
		Symbol     string `json:"symbol"`
		Binary     string `json:"binary"`
		SymbolName string `json:"symbolName"`
	} `json:"params"`
	Data struct {
		V string `json:"v"` //成交ID
		T int64  `json:"t"` //時間（撮合引擎撮合时间）
		P string `json:"p"` //價格
		Q string `json:"q"` //數量
		M bool   `json:"m"`
	} `json:"data"`
}
