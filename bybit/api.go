package bybit

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
)

type Client struct {
	baseURL          string // https://api-testnet.bybit.com/open-api/
	apiKey           string
	secretKey        string
	serverTimeOffset int64 // 时间偏差(ms)
	client           *http.Client
	debugMode        bool
}

func New(httpClient *http.Client, baseURL string, apiKey string, secretKey string, debugMode bool) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}
	return &Client{
		baseURL:   baseURL,
		apiKey:    apiKey,
		secretKey: secretKey,
		client:    httpClient,
		debugMode: debugMode,
	}
}

// SetCorrectServerTime 校正服务器时间
func (b *Client) SetCorrectServerTime() (err error) {
	var timeNow int64
	timeNow, err = b.GetServerTime()
	if err != nil {
		return
	}
	b.serverTimeOffset = timeNow - time.Now().UnixNano()/1e6
	return
}

// GetBalance Get Wallet Balance
// coin: BTC,EOS,XRP,ETH,USDT
func (b *Client) GetWalletBalance(coin string) (result Balance, err error) {
	var ret GetBalanceResult
	params := map[string]interface{}{}
	params["coin"] = coin
	_, err = b.SignedRequest(http.MethodGet, "v2/private/wallet/balance", params, &ret) // v2/private/wallet/balance
	if err != nil {
		return
	}
	switch coin {
	case "BTC":
		result = ret.Result.BTC
	case "ETH":
		result = ret.Result.ETH
	case "EOS":
		result = ret.Result.EOS
	case "XRP":
		result = ret.Result.XRP
	case "USDT":
		result = ret.Result.USDT
	}
	return
}

// GetLeverages 获取用户杠杆
func (b *Client) GetLeverages() (result map[string]LeverageItem, err error) {
	var r GetLeverageResult
	params := map[string]interface{}{}
	_, err = b.SignedRequest(http.MethodGet, "user/leverage", params, &r)
	if err != nil {
		return
	}
	result = r.Result
	return
}

// SetLeverage 设置杠杆
func (b *Client) SetLeverage(leverage int, symbol string) (err error) {
	var r BaseResult
	params := map[string]interface{}{}
	params["symbol"] = symbol
	params["buy_leverage"] = leverage
	params["sell_leverage"] = leverage
	_, err = b.SignedRequest(http.MethodPost, "private/linear/position/set-leverage", params, &r)
	if err != nil {
		return
	}
	return
}

// SetIsolated 设置杠杆
func (b *Client) SetIsolated(isIsolated bool, leverage int, symbol string) (err error) {
	var r BaseResult
	params := map[string]interface{}{}
	params["symbol"] = symbol
	params["is_isolated"] = isIsolated
	params["buy_leverage"] = leverage
	params["sell_leverage"] = leverage
	_, err = b.SignedRequest(http.MethodPost, "private/linear/position/set-leverage", params, &r)
	if err != nil {
		return
	}
	return
}

// GetPositions 获取我的仓位
func (b *Client) GetPositions() (result []Position, err error) {
	var r PositionListResult

	params := map[string]interface{}{}
	var resp []byte
	resp, err = b.SignedRequest(http.MethodGet, "private/linear/position/list", params, &r)
	if err != nil {
		return
	}
	if r.RetCode != 0 {
		err = fmt.Errorf("%v body: [%v]", r.RetMsg, string(resp))
		return
	}

	return
}

// GetPosition 获取我的仓位
func (b *Client) GetPosition(symbol string) (result []Position, err error) {
	var r GetPositionResult

	params := map[string]interface{}{}
	params["symbol"] = symbol
	var resp []byte
	resp, err = b.SignedRequest(http.MethodGet, "private/linear/position/list", params, &r)
	if err != nil {
		return
	}
	if r.RetCode != 0 {
		err = fmt.Errorf("%v body: [%v]", r.RetMsg, string(resp))
		return
	}
	result = r.Result
	return
}

// SetPositionMode 设置仓位模式
func (b *Client) SetPositionMode(symbol, mode string) (err error) {
	var r SetPositionModeResult

	params := map[string]interface{}{}
	params["symbol"] = symbol
	params["mode"] = mode
	var resp []byte
	resp, err = b.SignedRequest(http.MethodPost, "private/linear/position/switch-mode", params, &r)
	if err != nil {
		return
	}
	if r.RetCode != 0 {
		err = fmt.Errorf("%v body: [%v]", r.RetMsg, string(resp))
		return
	}
	return
}

func (b *Client) PublicRequest(method string, apiURL string, params map[string]interface{}, result interface{}) (resp []byte, err error) {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var p []string
	for _, k := range keys {
		p = append(p, fmt.Sprintf("%v=%v", k, params[k]))
	}

	param := strings.Join(p, "&")
	fullURL := b.baseURL + apiURL
	if param != "" {
		fullURL += "?" + param
	}
	//if b.debugMode {
	//	log.Printf("PublicRequest: %v", fullURL)
	//}
	var binBody = bytes.NewReader(make([]byte, 0))

	// get a http request
	var request *http.Request
	request, err = http.NewRequest(method, fullURL, binBody)
	if err != nil {
		return
	}
	//request.Header.Add(utils.ExKind, "bybit")
	var response *http.Response
	response, err = b.client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()

	resp, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	if b.debugMode {
		zap.S().Info("PublicRequest: %v", string(resp))
	}

	err = json.Unmarshal(resp, result)
	return
}

func (b *Client) SignedRequest(method string, apiURL string, params map[string]interface{}, result interface{}) (resp []byte, err error) {
	timestamp := time.Now().UnixNano()/1e6 + b.serverTimeOffset

	params["api_key"] = b.apiKey
	params["timestamp"] = timestamp
	params["recvWindow"] = recvWindow

	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var p []string
	for _, k := range keys {
		p = append(p, fmt.Sprintf("%v=%v", k, params[k]))
	}

	param := strings.Join(p, "&")
	signature := b.getSigned(param)
	param += "&sign=" + signature

	fullURL := b.baseURL + apiURL + "?" + param
	if b.debugMode {
		zap.S().Info("SignedRequest: %v", fullURL)
	}
	var binBody = bytes.NewReader(make([]byte, 0))

	// get a http request
	var request *http.Request
	request, err = http.NewRequest(method, fullURL, binBody)
	if err != nil {
		return
	}

	//request.Header.Add(utils.ExKind, "bybit")
	// 所有的签名接口都设置会不会有问题？
	request.Header.Add("Referer", "AntBot")

	var response *http.Response
	response, err = b.client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()

	resp, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	if b.debugMode {
		zap.S().Info("SignedRequest: %v", string(resp))
	}

	err = json.Unmarshal(resp, result)
	return
}

func (b *Client) getSigned(param string) string {
	sig := hmac.New(sha256.New, []byte(b.secretKey))
	sig.Write([]byte(param))
	signature := hex.EncodeToString(sig.Sum(nil))
	return signature
}
