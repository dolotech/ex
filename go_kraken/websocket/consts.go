package websocket

// URLs
const (
	ProdBaseURL_V2     = "wss://ws.kraken.com/v2"
	ProdBaseURL        = "wss://ws.kraken.com"
	AuthBaseURL        = "wss://ws-auth.kraken.com"
	SandboxBaseURL     = "wss://beta-ws.kraken.com"
	AuthSandboxBaseURL = "wss://beta-ws-auth.kraken.com"

	ProdBaseFuturesURL = "wss://futures.kraken.com/ws/v1"
)

// Available channels
const (
	ChanBook       = "book"
	ChanTrades     = "trade"
	ChanTicker     = "ticker"
	ChanCandles    = "ohlc"
	ChanSpread     = "spread"
	ChanOpenOrders = "openOrders"
	ChanOwnTrades  = "ownTrades"
	ChanAll        = "*"
)

const (
	FUTURES_Subscribed = "subscribed"
	FUTURES_Ticker     = "ticker"
	FUTURES_Info       = "info"
	FUTURES_Alert      = "alert"
	FUTURES_EventPing  = "ping"
	FUTURES_EventPong  = "pong"

	FUTURES_CANDLES_ = "candles_trade_"

	FUTURES_CANDLES_1M  = "candles_trade_1m"
	FUTURES_CANDLES_3M  = "candles_trade_3m"
	FUTURES_CANDLES_5M  = "candles_trade_5m"
	FUTURES_CANDLES_15M = "candles_trade_15m"
	FUTURES_CANDLES_30M = "candles_trade_30m"
	FUTURES_CANDLES_1H  = "candles_trade_1h"
	FUTURES_CANDLES_2H  = "candles_trade_2h"
	FUTURES_CANDLES_4H  = "candles_trade_4h"
	FUTURES_CANDLES_6H  = "candles_trade_6h"
	FUTURES_CANDLES_12H = "candles_trade_12h"
	FUTURES_CANDLES_1D  = "candles_trade_1d"

	FUTURES_CANDLES_1M_SNAPSHOT  = "candles_trade_1m_snapshot"
	FUTURES_CANDLES_3M_SNAPSHOT  = "candles_trade_3m_snapshot"
	FUTURES_CANDLES_5M_SNAPSHOT  = "candles_trade_5m_snapshot"
	FUTURES_CANDLES_15M_SNAPSHOT = "candles_trade_15m_snapshot"
	FUTURES_CANDLES_30M_SNAPSHOT = "candles_trade_30m_snapshot"
	FUTURES_CANDLES_1H_SNAPSHOT  = "candles_trade_1h_snapshot"
	FUTURES_CANDLES_2H_SNAPSHOT  = "candles_trade_2h_snapshot"
	FUTURES_CANDLES_4H_SNAPSHOT  = "candles_trade_4h_snapshot"
	FUTURES_CANDLES_6H_SNAPSHOT  = "candles_trade_6h_snapshot"
	FUTURES_CANDLES_12H_SNAPSHOT = "candles_trade_12h_snapshot"
	FUTURES_CANDLES_1D_SNAPSHOT  = "candles_trade_1d_snapshot"
)

// Events
const (
	EventSubscribe                  = "subscribe"
	EventUnsubscribe                = "unsubscribe"
	EventPing                       = "ping"
	EventPong                       = "pong"
	EventSystemStatus               = "systemStatus"
	EventSubscriptionStatus         = "subscriptionStatus"
	EventHeartbeat                  = "heartbeat"
	EventAddOrder                   = "addOrder"
	EventAddOrderStatus             = "addOrderStatus"
	EventCancelOrder                = "cancelOrder"
	EventCancelOrderStatus          = "cancelOrderStatus"
	EventCancelAll                  = "cancelAll"
	EventCancelAllStatus            = "cancelAllStatus"
	EventCancelAllOrdersAfter       = "cancelAllOrdersAfter"
	EventCancelAllOrdersAfterStatus = "cancelAllOrdersAfterStatus"
	EventEditOrder                  = "editOrder"
	EventEditOrderStatus            = "editOrderStatus"
)

// Intervals
const (
	Interval1     = 1
	Interval5     = 5
	Interval15    = 15
	Interval30    = 30
	Interval60    = 60
	Interval240   = 240
	Interval1440  = 1440
	Interval10080 = 10080
	Interval21600 = 21600
)

// Depth
const (
	Depth10   = 10
	Depth25   = 25
	Depth100  = 100
	Depth500  = 500
	Depth1000 = 1000
)

// Subscription Statuses
const (
	SubscriptionStatusError        = "error"
	SubscriptionStatusSubscribed   = "subscribed"
	SubscriptionStatusUnsubscribed = "unsubscribed"
)

// Trade sides for parsing
const (
	Buy  = "b"
	Sell = "s"
)

// Order types for parsing
const (
	Market = "m"
	Limit  = "l"
)

// Trade sides -
const (
	SideBuy  = "buy"
	SideSell = "sell"
)

// Order types -
const (
	OrderTypeMarket          = "market"
	OrderTypeLimit           = "limit"
	OrderTypeStopLoss        = "stop-loss"
	OrderTypeTakeProfit      = "take-profit"
	OrderTypeStopLossLimit   = "stop-loss-limit"
	OrderTypeTakeProfitLimit = "take-profit-limit"
	OrderTypeSettlePosition  = "settle-position"
)



// Pairs
const (
	ADACAD  = "ADA/CAD"
	ADAETH  = "ADA/ETH"
	ADAEUR  = "ADA/EUR"
	ADAUSD  = "ADA/USD"
	ADABTC  = "ADA/XBT"
	BCHEUR  = "BCH/EUR"
	BCHUSD  = "BCH/USD"
	BCHBTC  = "BCH/XBT"
	BSVEUR  = "BSV/EUR"
	BSVUSD  = "BSV/USD"
	BSVBTC  = "BSV/XBT"
	BTCEUR  = "XBT/EUR"
	BTCUSD  = "XBT/USD"
	BTCCAD  = "XBT/CAD"
	BTCJPY  = "XBT/JPY"
	BTCGBP  = "XBT/GBP"
	BTCUSDT = "XBT/USDT"
	DASHEUR = "DASH/EUR"
	DASHUSD = "DASH/USD"
	DASHBTC = "DASH/XBT"
	DOGEBTC = "XDG/XBT"
	DOTEUR  = "DOT/EUR"
	DOTUSD  = "DOT/USD"
	EOSETH  = "EOS/ETH"
	EOSEUR  = "EOS/EUR"
	EOSUSD  = "EOS/USD"
	EOSBTC  = "EOS/XBT"
	ETCETH  = "ETC/ETH"
	ETCEUR  = "ETC/EUR"
	ETCUSD  = "ETC/USD"
	ETCBTC  = "ETC/XBT"
	ETHCAD  = "ETH/CAD"
	ETHEUR  = "ETH/EUR"
	ETHUSD  = "ETH/USD"
	ETHBTC  = "ETH/XBT"
	ETHJPY  = "ETH/JPY"
	ETHGBP  = "ETH/GBP"
	ETHUSDT = "ETH/USDT"
	GNOETH  = "GNO/ETH"
	GNOEUR  = "GNO/EUR"
	GNOUSD  = "GNO/USD"
	GNOBTC  = "GNO/XBT"
	LTCEUR  = "LTC/EUR"
	LTCUSD  = "LTC/USD"
	LTCBTC  = "LTC/XBT"
	LTCUSDT = "LTC/USDT"
	MLNETH  = "MLN/ETH"
	MLNBTC  = "MLN/XBT"
	QTUMCAD = "QTUM/CAD"
	QTUMETH = "QTUM/ETH"
	QTUMEUR = "QTUM/EUR"
	QTUMUSD = "QTUM/USD"
	QTUMBTC = "QTUM/XBT"
	REPETH  = "REP/ETH"
	REPEUR  = "REP/EUR"
	REPUSD  = "REP/USD"
	REPBTC  = "REP/XBT"
	USDTUSD = "USDT/USD"
	XLMEUR  = "XLM/EUR"
	XLMUSD  = "XLM/USD"
	XLMBTC  = "XLM/XBT"
	XMREUR  = "XMR/EUR"
	XMRUSD  = "XMR/USD"
	XMRBTC  = "XMR/XBT"
	XRPCAD  = "XRP/CAD"
	XRPEUR  = "XRP/EUR"
	XRPJPY  = "XRP/JPY"
	XRPUSD  = "XRP/USD"
	XRPBTC  = "XRP/XBT"
	XTZCAD  = "XTZ/CAD"
	XTZETH  = "XTZ/ETH"
	XTZEUR  = "XTZ/EUR"
	XTZUSD  = "XTZ/USD"
	XTZBTC  = "XTZ/XBT"
	ZECEUR  = "ZEC/EUR"
	ZECJPY  = "ZEC/JPY"
	ZECUSD  = "ZEC/USD"
)

// Statuses
const (
	StatusOK    = "ok"
	StatusError = "error"
)
