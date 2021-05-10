package mcap

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/go-querystring/query"
)

const (
	basePath   = "https://marketcap.backend.currency.com"
	apiName    = "/api"
	apiVersion = "/v1"
)

type Client struct {
	Client  *http.Client
	BaseURL *url.URL
}

func (c *Client) GetCandles(cr CandlesRequest) (candles Candles, err error) {
	var rowData = candlesRowData{}
	var candle Candle

	err = c.doRequest("/candles", cr, &rowData)
	if err != nil {
		return
	}

	for _, bar := range rowData {
		candle, err = prepareCandle(bar)
		if err != nil {
			return
		}
		candles = append(candles, candle)
	}
	return
}

func (c *Client) GetAssets() (listOfAssets Assets, err error) {
	err = c.doRequest("/assets", nil, &listOfAssets)
	return
}

func (c *Client) GetOrderbook(obr OrderbookRequest) (ob Orderbook, err error) {
	err = c.doRequest("/orderbook", obr, &ob)
	return
}

func (c *Client) GetSummary() (md MarketData, err error) {
	err = c.doRequest("/summary", nil, &md)
	return
}

func (c *Client) GetTicker() (t Ticker, err error) {
	err = c.doRequest("/ticker", nil, &t)
	return
}

func (c *Client) GetTrades(tr TradesRequest) (t Trades, err error) {
	err = c.doRequest("/trades", tr, &t)
	return
}

func prepareCandle(bar []interface{}) (candle Candle, err error) {
	timeInt, err := strconv.ParseInt(strconv.FormatFloat(bar[0].(float64), 'f', 0, 64), 10, 64)
	if err != nil {
		return
	}
	tm := time.Unix(timeInt/1000, 0)

	o, err := strconv.ParseFloat(fmt.Sprintf("%v", bar[1]), 64)
	if err != nil {
		return
	}
	h, err := strconv.ParseFloat(fmt.Sprintf("%v", bar[2]), 64)
	if err != nil {
		return
	}
	l, err := strconv.ParseFloat(fmt.Sprintf("%v", bar[3]), 64)
	if err != nil {
		return
	}
	c, err := strconv.ParseFloat(fmt.Sprintf("%v", bar[4]), 64)
	if err != nil {
		return
	}

	return Candle{Time: tm, Open: o, High: h, Low: l, Close: c}, nil
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURL, _ := url.Parse(basePath + apiName + apiVersion)
	c := &Client{Client: httpClient, BaseURL: baseURL}
	return c
}

func (c *Client) doRequest(endpoint string, values interface{}, results interface{}) error {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL.String()+endpoint, nil)
	if err != nil {
		return err
	}

	params, err := query.Values(values)
	if err != nil {
		return err
	}

	req.URL.RawQuery = params.Encode()
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(params.Encode())))

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	err = checkResponse(resp)
	if err != nil {
		return err
	}
	return json.NewDecoder(resp.Body).Decode(&results)
}

func checkResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{
		Response: r,
		Status:   r.StatusCode,
		ErrorMsg: "something went wrong",
	}

	data, err := ioutil.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		err := json.Unmarshal(data, errorResponse)
		if err != nil {
			errorResponse.Message = string(data)
		}
	}
	return errorResponse
}

type CandlesRequest struct {
	EndTime   int64  `url:"endTime,omitempty"`
	Interval  string `url:"interval"`
	Limit     int    `url:"limit,omitempty"`
	StartTime int64  `url:"startTime,omitempty"`
	Symbol    string `url:"symbol"`
}

type candlesRowData [][]interface{}

type Candles []Candle

type Candle struct {
	Time  time.Time
	High  float64
	Low   float64
	Open  float64
	Close float64
}

type Assets map[string]Asset

type Asset struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	CanWithdraw bool    `json:"can_withdraw"`
	CanDeposit  bool    `json:"can_deposit"`
	MinWithdraw float64 `json:"min_withdraw"`
	MaxWithdraw float64 `json:"max_withdraw"`
	MakerFee    float64 `json:"maker_fee"`
	TakerFee    float64 `json:"taker_fee"`
}

type OrderbookRequest struct {
	Depth  int    `url:"depth,omitempty"`
	Level  int    `url:"level,omitempty"`
	Symbol string `url:"symbol"`
}

type Orderbook struct {
	Timestamp int64       `json:"timestamp"`
	Asks      [][]float64 `json:"asks"`
	Bids      [][]float64 `json:"bids"`
}

type MarketData struct {
	Message string                `json:"msg"`
	Data    map[string]MarketProp `json:"data"`
}
type MarketProp struct {
	ID            int64  `json:"id"`
	BaseVolume    string `json:"baseVolume"`
	High24Hr      string `json:"high24hr"`
	HighestBid    string `json:"highestBid"`
	IsFrozen      string `json:"isFrozen"`
	Last          string `json:"last"`
	Low24Hr       string `json:"low24hr"`
	LowestAsk     string `json:"lowestAsk"`
	PercentChange string `json:"percentChange"`
	QuoteVolume   string `json:"quoteVolume"`
}

type Ticker map[string]PriceChange

type PriceChange struct {
	BaseCurrency         string  `json:"base_currency"`
	BaseVolume           float64 `json:"base_volume"`
	Description          string  `json:"description"`
	HighestBidPrice      float64 `json:"highest_bid_price"`
	IsFrozen             bool    `json:"isFrozen"`
	LastPrice            float64 `json:"last_price"`
	LowestAskPrice       float64 `json:"lowest_ask_price"`
	Past24HrsHighPrice   float64 `json:"past_24hrs_high_price"`
	Past24HrsLowPrice    float64 `json:"past_24hrs_low_price"`
	Past24HrsPriceChange float64 `json:"past_24hrs_price_change"`
	QuoteVolume          float64 `json:"quote_volume"`
	QuoteCurrency        string  `json:"quote_currency"`
}

type Trades []struct {
	TradeID        int64   `json:"tradeID"`
	Price          float64 `json:"price"`
	BaseVolume     float64 `json:"base_volume"`
	QuoteVolume    float64 `json:"quote_volume"`
	TradeTimestamp int64   `json:"trade_timestamp"`
	Type           string  `json:"type"`
}

type TradesRequest struct {
	Symbol string `url:"symbol"`
	Type   string `url:"type,omitempty"`
}

type ErrorResponse struct {
	// HTTP response that caused this error
	Response *http.Response `json:"-"`
	// Error message
	Timestamp int64  `json:"timestamp"`
	Status    int    `json:"status"`
	ErrorMsg  string `json:"error"`
	Message   string `json:"message"`
	Path      string `json:"path"`
	Msg       string `json:"msg"`
	Code      int    `json:"code"`
}

func (r *ErrorResponse) Error() string {
	msg := r.Message
	if r.Code < 0 {
		msg = r.Msg
	}
	return fmt.Sprintf("%v %v: [%s] %v", r.Response.Request.Method, r.Response.Request.URL, r.Response.Status, msg)
}
