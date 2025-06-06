package bybit

import "net/http"

// 獲取錢包余額
func (b *Client) GetSpotAccount() (Balances, error) {
	var ret AccountResult
	params := map[string]interface{}{}
	_, err := b.PublicRequest(http.MethodGet, "spot/v1/account", params, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Result.Balances, nil
}

type Balances []BalanceVo

type AccountResult struct {
	RetCode int         `json:"ret_code"`
	RetMsg  string      `json:"ret_msg"`
	ExtCode interface{} `json:"ext_code"`
	ExtInfo interface{} `json:"ext_info"`
	Result  struct {
		Balances `json:"balances"`
	} `json:"result"`
}

type BalanceVo struct {
	Coin     string `json:"coin"`
	CoinId   string `json:"coinId"`
	CoinName string `json:"coinName"`
	Total    string `json:"total"`
	Free     string `json:"free"`
	Locked   string `json:"locked"`
}

// 獲取錢包余額
func (b *Client) GetFuturesAccount() (map[string]interface{}, error) {
	var ret FuturesAccountResult
	params := map[string]interface{}{}
	_, err := b.PublicRequest(http.MethodGet, "v2/private/wallet/balance", params, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Result.Result, nil
}

type FuturesAccountResult struct {
	RetCode int    `json:"ret_code"`
	RetMsg  string `json:"ret_msg"`
	ExtCode string `json:"ext_code"`
	ExtInfo string `json:"ext_info"`
	Result  struct {
		Result map[string]interface{}
	} `json:"result"`
	TimeNow          string `json:"time_now"`
	RateLimitStatus  int    `json:"rate_limit_status"`
	RateLimitResetMs int64  `json:"rate_limit_reset_ms"`
	RateLimit        int    `json:"rate_limit"`
}

type BalancesFutures struct {
	Equity           int     `json:"equity"`
	AvailableBalance float64 `json:"available_balance"`
	UsedMargin       float64 `json:"used_margin"`
	OrderMargin      float64 `json:"order_margin"`
	PositionMargin   int     `json:"position_margin"`
	OccClosingFee    int     `json:"occ_closing_fee"`
	OccFundingFee    int     `json:"occ_funding_fee"`
	WalletBalance    int     `json:"wallet_balance"`
	RealisedPnl      int     `json:"realised_pnl"`
	UnrealisedPnl    int     `json:"unrealised_pnl"`
	CumRealisedPnl   int     `json:"cum_realised_pnl"`
	GivenCash        int     `json:"given_cash"`
	ServiceCash      int     `json:"service_cash"`
}
