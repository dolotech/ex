package websocket

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// Update - notification from channel or events
type Update struct {
	ChannelID   int64
	Data        interface{}
	ChannelName string
	Pair        string
	Sequence    Seq
}

// Message - data structure of default Kraken WS update
type Message struct {
	ChannelID   int64
	Data        json.RawMessage
	ChannelName string
	Pair        string
	Sequence    Seq
}

func (msg Message) toUpdate(data interface{}) Update {
	return Update{
		ChannelID:   msg.ChannelID,
		Data:        data,
		ChannelName: msg.ChannelName,
		Pair:        msg.Pair,
		Sequence:    msg.Sequence,
	}
}

// Seq -
type Seq struct {
	Value int64 `json:"sequence"`
}

// UnmarshalJSON - unmarshal update
func (msg *Message) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw) < 3 {
		return fmt.Errorf("invalid data length: %#v", raw)
	}

	if len(raw) == 5 {
		// order book can have 2 data objects
		// one for the new asks and one for the new bids
		// see https://docs.kraken.com/websockets/

		// the array is [channelid, ask, bid, channel, pair]
		ask := raw[1]
		bid := raw[2]

		// ask and bid can be merged into a single object as the keys are distinct
		if ask[len(ask)-1] != '}' || bid[0] != '{' {
			// not a bid/ask pair
			return fmt.Errorf("invalid data length/payload: %v", raw)
		}

		// merge ask + bid
		merged := make([]byte, 0, len(ask)+len(bid)-1)
		merged = append(merged, ask[0:len(ask)-1]...)
		merged = append(merged, ',')
		merged = append(merged, bid[1:]...)

		// reencode
		data, _ = json.Marshal([]json.RawMessage{
			raw[0], merged, raw[3], raw[4],
		})
	}

	body := make([]interface{}, 0)
	if len(raw) == 3 {
		body = append(body, &msg.Data, &msg.ChannelName, &msg.Sequence)
	} else {
		body = append(body, &msg.ChannelID, &msg.Data, &msg.ChannelName, &msg.Pair)
	}

	return json.Unmarshal(data, &body)
}

// TickerUpdate - data structure for ticker update
type TickerUpdate struct {
	Ask                Level         `json:"a"`
	Bid                Level         `json:"b"`
	Close              DecimalValues `json:"c"`
	Volume             DecimalValues `json:"v"`
	VolumeAveragePrice DecimalValues `json:"p"`
	TradeVolume        IntValues     `json:"t"`
	Low                DecimalValues `json:"l"`
	High               DecimalValues `json:"h"`
	Open               DecimalValues `json:"o"`
}

// Level -
type Level struct {
	Price          json.Number
	Volume         json.Number
	WholeLotVolume int
}

// UnmarshalJSON - unmarshal ticker update
func (l *Level) UnmarshalJSON(data []byte) error {
	raw := []interface{}{&l.Price, &l.WholeLotVolume, &l.Volume}
	return json.Unmarshal(data, &raw)
}

// DecimalValues - data structure for decimal ticker data
type DecimalValues struct {
	Today  json.Number
	Last24 json.Number
}

// UnmarshalJSON - unmarshal ticker update
func (v *DecimalValues) UnmarshalJSON(data []byte) error {
	raw := []interface{}{&v.Today, &v.Last24}
	return json.Unmarshal(data, &raw)
}

// IntValues - data structure for int ticker data
type IntValues struct {
	Today  int64
	Last24 int64
}

// UnmarshalJSON - unmarshal ticker update
func (v *IntValues) UnmarshalJSON(data []byte) error {
	raw := []interface{}{&v.Today, &v.Last24}
	return json.Unmarshal(data, &raw)
}

// Candle -
type Candle struct {
	Time      json.Number
	EndTime   json.Number
	Open      json.Number
	High      json.Number
	Low       json.Number
	Close     json.Number
	VolumeWAP json.Number
	Volume    json.Number
	Count     int64
}

// UnmarshalJSON - unmarshal candle update
func (c *Candle) UnmarshalJSON(data []byte) error {
	raw := []interface{}{&c.Time, &c.EndTime, &c.Open, &c.High, &c.Low, &c.Close, &c.VolumeWAP, &c.Volume, &c.Count}
	return json.Unmarshal(data, &raw)
}

// Trade - data structure for trade update
type Trade struct {
	Price     json.Number
	Volume    json.Number
	Time      json.Number
	Side      string
	OrderType string
	Misc      string
}

// UnmarshalJSON - unmarshal candle update
func (t *Trade) UnmarshalJSON(data []byte) error {
	raw := []interface{}{&t.Price, &t.Volume, &t.Time, &t.Side, &t.OrderType, &t.Misc}
	return json.Unmarshal(data, &raw)
}

// Spread - data structure for spread update
type Spread struct {
	Ask       json.Number
	Bid       json.Number
	AskVolume json.Number
	BidVolume json.Number
	Time      json.Number
}

// UnmarshalJSON - unmarshal candle update
func (s *Spread) UnmarshalJSON(data []byte) error {
	raw := []interface{}{&s.Bid, &s.Ask, &s.Time, &s.AskVolume, &s.BidVolume, &s.Time}
	return json.Unmarshal(data, &raw)
}

// OrderBookItem - data structure for order book item
type OrderBookItem struct {
	Price     json.Number
	Volume    json.Number
	Time      json.Number
	Republish bool
}

// UnmarshalJSON - unmarshal candle update
func (obi *OrderBookItem) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw) < 3 {
		return errors.Errorf("Invalid order book item: %s", string(data))
	}

	obi.Republish = len(raw) == 4

	if err := json.Unmarshal(raw[0], &obi.Price); err != nil {
		return err
	}

	if err := json.Unmarshal(raw[1], &obi.Volume); err != nil {
		return err
	}

	return json.Unmarshal(raw[2], &obi.Time)
}

// OrderBookUpdate - data structure for order book update
type OrderBookUpdate struct {
	Asks       []OrderBookItem
	Bids       []OrderBookItem
	CheckSum   string
	IsSnapshot bool
}

// UnmarshalJSON - unmarshal candle update
func (obu *OrderBookUpdate) UnmarshalJSON(data []byte) error {
	body := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &body); err != nil {
		return err
	}

	for key, raw := range body {
		if len(key) == 2 {
			obu.IsSnapshot = true
		}

		switch key[0] {
		case 'a':
			if err := json.Unmarshal(raw, &obu.Asks); err != nil {
				return err
			}
		case 'b':
			if err := json.Unmarshal(raw, &obu.Bids); err != nil {
				return err
			}
		case 'c':
			if err := json.Unmarshal(raw, &obu.CheckSum); err != nil {
				return err
			}
		}
	}

	return nil
}

// OwnTrade - Own trades.
type OwnTrade struct {
	Cost      json.Number `json:"cost"`
	Fee       json.Number `json:"fee"`
	Margin    json.Number `json:"margin"`
	OrderID   string      `json:"ordertxid"`
	OrderType string      `json:"ordertype"`
	Pair      string      `json:"pair"`
	PosTxID   string      `json:"postxid"`
	Price     json.Number `json:"price"`
	Time      json.Number `json:"time"`
	Type      string      `json:"type"`
	Vol       json.Number `json:"vol"`
	UserRef   json.Number `json:"userref"`
}

// OpenOrderDescr -
type OpenOrderDescr struct {
	Close     string      `json:"close"`
	Leverage  string      `json:"leverage"`
	Order     string      `json:"order"`
	Ordertype string      `json:"ordertype"`
	Pair      string      `json:"pair"`
	Price     json.Number `json:"price"`
	Price2    json.Number `json:"price2"`
	Type      string      `json:"type"`
}

// OpenOrder -
type OpenOrder struct {
	Cost       json.Number    `json:"cost"`
	Descr      OpenOrderDescr `json:"descr"`
	Fee        json.Number    `json:"fee"`
	LimitPrice json.Number    `json:"limitprice"`
	Misc       string         `json:"misc"`
	Oflags     string         `json:"oflags"`
	OpenTime   json.Number    `json:"opentm"`
	StartTime  json.Number    `json:"starttm"`
	ExpireTime json.Number    `json:"expiretm"`
	Price      json.Number    `json:"price"`
	Refid      string         `json:"refid"`
	Status     string         `json:"status"`
	StopPrice  json.Number    `json:"stopprice"`
	UserRef    int64          `json:"userref"`
	Vol        json.Number    `json:"vol,string"`
	VolExec    json.Number    `json:"vol_exec"`
}

// OwnTradesUpdate -
type OwnTradesUpdate []map[string]OwnTrade

// OpenOrdersUpdate -
type OpenOrdersUpdate []map[string]OpenOrder

// {"time":1721213867738,"product_id":"PF_SOLUSD","funding_rate":0.001318439968712204,"funding_rate_prediction":0.00193406121869099,"relative_funding_rate":8.060811111111e-6,"relative_funding_rate_prediction":0.000011884326388889,"next_funding_rate_time":1721214000000,"leverage":"50x","premium":0.1,"feed":"ticker","bid":162.74,"ask":162.75,"bid_size":23.4,"ask_size":9.5,"volume":378480.79,"dtm":0,"index":162.6634,"last":162.74,"change":3.87,"suspended":false,"tag":"perpetual","pair":"SOL:USD","openInterest":156019.07,"markPrice":162.75734404771,"maturityTime":0,"post_only":false,"volumeQuote":60666084.9734,"open":156.67,"high":163.89,"low":154.57}
type FuturesTicker struct {
	Time                          int64   `json:"time"`
	ProductId                     string  `json:"product_id"`
	FundingRate                   float64 `json:"funding_rate"`
	FundingRatePrediction         float64 `json:"funding_rate_prediction"`
	RelativeFundingRate           float64 `json:"relative_funding_rate"`
	RelativeFundingRatePrediction float64 `json:"relative_funding_rate_prediction"`
	NextFundingRateTime           int64   `json:"next_funding_rate_time"`
	Leverage                      string  `json:"leverage"`
	Premium                       float64 `json:"premium"`
	Feed                          string  `json:"feed"`
	Bid                           float64 `json:"bid"`
	Ask                           float64 `json:"ask"`
	BidSize                       float64 `json:"bid_size"`
	AskSize                       float64 `json:"ask_size"`
	Volume                        float64 `json:"volume"`
	Dtm                           int     `json:"dtm"`
	Index                         float64 `json:"index"`
	Last                          float64 `json:"last"`
	Change                        float64 `json:"change"`
	Suspended                     bool    `json:"suspended"`
	Tag                           string  `json:"tag"`
	Pair                          string  `json:"pair"`
	OpenInterest                  float64 `json:"openInterest"`
	MarkPrice                     float64 `json:"markPrice"`
	MaturityTime                  int     `json:"maturityTime"`
	PostOnly                      bool    `json:"post_only"`
	VolumeQuote                   float64 `json:"volumeQuote"`
	Open                          float64 `json:"open"`
	High                          float64 `json:"high"`
	Low                           float64 `json:"low"`
}

type FuturesAlert struct {
	Event   string `json:"event"`
	Message string `json:"message"`
}
type FuturesInfo struct {
	Event   string `json:"event"`
	Version int    `json:"version"`
}

type FuturesCandle struct {
	Feed   string `json:"feed"`
	Candle struct {
		Time   int64  `json:"time"`
		Open   string `json:"open"`
		High   string `json:"high"`
		Low    string `json:"low"`
		Close  string `json:"close"`
		Volume string `json:"volume"`
	} `json:"candle"`
	ProductId string `json:"product_id"`
}

type FuturesSubscribed struct {
	Event      string   `json:"event"`
	Feed       string   `json:"feed"`
	ProductIds []string `json:"product_ids"`
}
