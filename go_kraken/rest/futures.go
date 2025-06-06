package rest

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// https://api.futures.kraken.com/derivatives/api/v3
// wss://api.futures.kraken.com/ws/v1
const (
	APIFuturesBase = "https://futures.kraken.com/derivatives"

	APIFuturesV3 = "/api/v3"
	APIFuturesV4 = "/api/v4"
)

func (api *Kraken) Authentication(endPoint string, postData string) (nonce string, authent string) {
	nonce = strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	input := postData + nonce + endPoint
	hash := sha256.Sum256([]byte(input))
	macKey, _ := base64.StdEncoding.DecodeString(api.secret)
	mac := hmac.New(sha512.New, macKey)
	mac.Write(hash[:])
	authent = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return
}

func (api *Kraken) prepareFuturesRequest(method string, data url.Values, retType interface{}) error {
	if data == nil {
		data = url.Values{}
	}
	endPoint := fmt.Sprintf("%s/%s", APIFuturesV3, method)
	requestURL := fmt.Sprintf("%s%s", APIFuturesBase, endPoint)
	reqData := data.Encode()
	req, err := http.NewRequest("GET", requestURL, strings.NewReader(reqData))
	if err != nil {
		return errors.Wrap(err, "error during request creation")
	}

	req.Header.Add("Accept", "application/json")
	if len(api.key) > 0 {
		nonce, authent := api.Authentication(endPoint, reqData)
		req.Header.Add("APIKey", api.key)
		req.Header.Add("Nonce", nonce)
		req.Header.Add("Authent", authent)
	}

	resp, err := api.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "error during request execution")
	}
	defer resp.Body.Close()
	return api.parseFuturessponse(resp, retType)
}

func (api *Kraken) parseFuturessponse(response *http.Response, retType interface{}) error {
	if response.StatusCode != 200 {
		return errors.Errorf("error during response parsing: invalid status code %d", response.StatusCode)
	}

	if response.Body == nil {
		return errors.New("error during response parsing: can not read response body")
	}

	if body, err := io.ReadAll(response.Body); err != nil {
		return err
	} else {
		jsonRes := &ResponseDerivatives{}
		if err := json.Unmarshal(body, jsonRes); err != nil {
			return err
		}
		if jsonRes.Result != "success" {
			return fmt.Errorf("%s", jsonRes.Error)
		}
		if err := json.Unmarshal(body, retType); err != nil {
			return err
		}
	}

	return nil
}
