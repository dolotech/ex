package bybit

import (
	"net/http"
	"sort"
	"strconv"
	"time"
)

// GetServerTime Get server time.
func (b *Client) GetServerTime() (timeNow int64, err error) {
	params := map[string]interface{}{}
	var ret BaseResult
	_, err = b.PublicRequest(http.MethodGet, "v2/public/time", params, &ret)
	if err != nil {
		return
	}
	var t float64
	t, err = strconv.ParseFloat(ret.TimeNow, 64)
	if err != nil {
		return
	}
	timeNow = int64(t * 1000)
	return
}

// GetOrderBook Get the orderbook
// 正反向合约通用
func (b *Client) GetOrderBook(symbol string) (result OrderBook, err error) {
	var ret GetOrderBookResult
	params := map[string]interface{}{}
	params["symbol"] = symbol
	_, err = b.PublicRequest(http.MethodGet, "v2/public/orderBook/L2", params, &ret)
	if err != nil {
		return
	}

	for _, v := range ret.Result {
		if v.Side == "Sell" {
			result.Asks = append(result.Asks, Item{
				Price: v.Price,
				Size:  v.Size,
			})
		} else if v.Side == "Buy" {
			result.Bids = append(result.Bids, Item{
				Price: v.Price,
				Size:  v.Size,
			})
		}
	}

	sort.Slice(result.Asks, func(i, j int) bool {
		return result.Asks[i].Price < result.Asks[j].Price
	})

	sort.Slice(result.Bids, func(i, j int) bool {
		return result.Bids[i].Price > result.Bids[j].Price
	})

	var timeNow float64
	timeNow, err = strconv.ParseFloat(ret.TimeNow, 64) // 1582011750.433202
	if err != nil {
		return
	}
	result.Time = time.Unix(0, int64(timeNow*1e9))
	return
}

// GetKLine
// https://bybit-exchange.github.io/docs/inverse/#t-httprequest-2
// interval: 1 3 5 15 30 60 120 240 360 720 "D" "M" "W" "Y"
// from: From timestamp in seconds
// limit: Limit for data size per page, max size is 200. Default as showing 200 pieces of data per page
func (b *Client) GetKLine(symbol string, interval string, from int64, limit int) (result []OHLC, err error) {
	var ret GetKlineResult
	params := map[string]interface{}{}
	params["symbol"] = symbol
	params["interval"] = interval
	params["from"] = from
	if limit > 0 {
		params["limit"] = limit
	}
	_, err = b.PublicRequest(http.MethodGet, "v2/public/kline/list", params, &ret)
	if err != nil {
		return
	}
	result = ret.Result
	return
}

func (b *Client) GetTickers() (result []Ticker, err error) {
	// https://api-testnet.bybit.com/v2/public/tickers
	var ret GetTickersResult
	params := map[string]interface{}{}
	_, err = b.PublicRequest(http.MethodGet, "v2/public/tickers", params, &ret)
	if err != nil {
		return
	}
	result = ret.Result
	return
}

// from: From ID. Default: return latest data
// limit: Number of results. Default 500; max 1000
func (b *Client) GetTradingRecords(symbol string, from int64, limit int) (result []TradingRecord, err error) {
	var ret GetTradingRecordsResult
	params := map[string]interface{}{}
	params["symbol"] = symbol
	if from > 0 {
		params["from"] = from
	}
	if limit > 0 {
		params["limit"] = limit
	}
	_, err = b.PublicRequest(http.MethodGet, "v2/public/trading-records", params, &ret)
	if err != nil {
		return
	}
	result = ret.Result
	return
}

func (b *Client) GetSymbols() (result []SymbolInfo, err error) {
	var ret GetSymbolsResult
	params := map[string]interface{}{}
	_, err = b.PublicRequest(http.MethodGet, "v2/public/symbols", params, &ret)
	if err != nil {
		return
	}
	result = ret.Result
	return
}

// 现货交易对信息
func (b *Client) GetSpotSymbols() ([]Symbol, error) {
	var ret SymbolsResult
	params := map[string]interface{}{}
	_, err := b.PublicRequest(http.MethodGet, "spot/v1/symbols", params, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Result, nil
}

type SymbolsResult struct {
	RetCode int         `json:"ret_code"`
	RetMsg  string      `json:"ret_msg"`
	ExtCode interface{} `json:"ext_code"`
	ExtInfo interface{} `json:"ext_info"`
	Result  []Symbol    `json:"result"`
}

// 交易对信息
type Symbol struct {
	Name              string `json:"name"`              // 幣對名稱
	Alias             string `json:"alias"`             // 幣對別名
	BaseCurrency      string `json:"baseCurrency"`      // base幣種
	QuoteCurrency     string `json:"quoteCurrency"`     // quote幣種
	BasePrecision     string `json:"basePrecision"`     // base幣種精度
	QuotePrecision    string `json:"quotePrecision"`    // quote幣種精度
	MinTradeQuantity  string `json:"minTradeQuantity"`  // 最小訂單數量
	MinTradeAmount    string `json:"minTradeAmount"`    // 最小訂單額
	MinPricePrecision string `json:"minPricePrecision"` // 最小價格精度
	MaxTradeQuantity  string `json:"maxTradeQuantity"`  // 最大成交量
	MaxTradeAmount    string `json:"maxTradeAmount"`    // 最大成交額
	Category          int    `json:"category"`          // symbol 所在分區:1主類別
}
