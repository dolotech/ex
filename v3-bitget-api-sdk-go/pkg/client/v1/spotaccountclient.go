package v1

import (
	"github.com/339-Labs/v3-bitget-api-sdk-go/common"
	"github.com/339-Labs/v3-bitget-api-sdk-go/config"
	"github.com/339-Labs/v3-bitget-api-sdk-go/internal"
)

type SpotAccountClient struct {
	BitgetRestClient *common.BitgetRestClient
}

func (p *SpotAccountClient) Init(config *config.BitgetConfig) *SpotAccountClient {
	p.BitgetRestClient = new(common.BitgetRestClient).Init(config)
	return p
}

func (p *SpotAccountClient) Info() (string, error) {
	params := internal.NewParams()
	resp, err := p.BitgetRestClient.DoGet("/api/spot/v1/account/getInfo", params)
	return resp, err
}

func (p *SpotAccountClient) Assets(params map[string]string) (string, error) {
	resp, err := p.BitgetRestClient.DoGet("/api/spot/v1/account/assets-lite", params)
	return resp, err
}

func (p *SpotAccountClient) Bills(params map[string]string) (string, error) {
	resp, err := p.BitgetRestClient.DoGet("/api/spot/v1/account/bills", params)
	return resp, err
}

func (p *SpotAccountClient) TransferRecords(params map[string]string) (string, error) {
	resp, err := p.BitgetRestClient.DoGet("/api/spot/v1/account/transferRecords", params)
	return resp, err
}
