package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

/*
	{Side:long Symbol:PF_DOGEUSD Price:0.125601 FillTime:2024-07-16 08:13:28.188 +0000 UTC Size:30 UnrealizedFunding:-1.3905733905161844e-05 PnlCurrency:USD}
	{Side:long Symbol:PF_WLDUSD Price:1.8298 FillTime:2024-07-16 08:13:28.188 +0000 UTC Size:2 UnrealizedFunding:4.141589343418913e-06 PnlCurrency:USD}
*/
// 获取账户下的合约仓位
func (api *Kraken) GetFuturesPositions() (OpenPositions, error) {
	response := OpenPositions{}

	err := api.prepareFuturesRequest("openpositions", nil, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// https://futures.kraken.com/derivatives/api/v4/charts/trade/PF_XBTUSD/15m?from=1721110189&to=1721214529
func (api *Kraken) GetFuturesCandles(symbol, period string, from, to int64) (FuturesCandles, error) {
	var ret FuturesCandles
	requestURL := fmt.Sprintf("%s/%s/%s/%s", APIFuturesBase, "api/v4/charts/trade", symbol, period)
	data := url.Values{}

	data.Add("from", fmt.Sprintf("%d", from))
	data.Add("to", fmt.Sprintf("%d", to))
	req, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", requestURL, data.Encode()), nil)
	if err != nil {
		return ret, err
	}
	req.Header.Add("Accept", "application/json")

	resp, err := api.client.Do(req)
	if err != nil {
		return ret, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ret, fmt.Errorf("error during response parsing: invalid status code %d symbol:%s period:%s", resp.StatusCode, symbol, period)
	}

	if resp.Body == nil {
		return ret, fmt.Errorf("error during response parsing: can not read response body")
	}

	if body, err := io.ReadAll(resp.Body); err != nil {
		return ret, err
	} else {
		if err := json.Unmarshal(body, &ret); err != nil {
			return ret, err
		}
	}
	return ret, nil
}
