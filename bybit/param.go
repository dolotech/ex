package bybit

const (
	SideBuy  = "Buy"
	SideSell = "Sell"
)

const (
	OrderTypeLimit  = "Limit"  // 限價單
	OrderTypeMarket = "Market" // 市價單
)

const (
	PositionIdxSingle   = iota // 0-單向持倉
	PositionIdxBothBuy         // 1-雙向持倉Buy
	PositionIdxBothSell        // 2-雙向持倉Sell
)

const (
	TimeInForceGoodTillCancel    = "GoodTillCancel"     // 一直有效至取消
	TimeInForceImmediateOrCancel = "ImmediateOrCancel " // 立即成交或取消
	TimeInForceFillOrKill        = "FillOrKill"         // 完全成交或取消
	TimeInForcePostOnly          = "PostOnly"           // 被動委托
)

const (
	PositionModeSingle = "MergedSingle"
	PositionModeBoth   = "BothSide"
)

const (
	recvWindow = 20000
)

type CreateOrderParam struct {
	Side           string
	OrderType      string
	Price          float64
	Qty            float64
	TimeInForce    string
	ReduceOnly     bool
	CloseOnTrigger bool
	Symbol         string
	OrderLinkID    string
	TakeProfit     float64
	StopLoss       float64
	TpTriggerBy    string
	SlTriggerBy    string
	PositionIdx    uint8
}
