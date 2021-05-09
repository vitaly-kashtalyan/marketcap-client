package mcap

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
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
	var rowData = CandlesRowData{}
	var candle Candle

	err = c.DoRequest("candles", cr, &rowData)
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

func (c *Client) DoRequest(endpoint string, values interface{}, results interface{}) error {
	c.BaseURL.Path = path.Join(c.BaseURL.Path, endpoint)
	req, err := http.NewRequest(http.MethodGet, c.BaseURL.String(), nil)
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

	err = CheckResponse(resp)
	if err != nil {
		return err
	}

	return json.NewDecoder(resp.Body).Decode(&results)
}

func CheckResponse(r *http.Response) error {
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

type CandlesRowData [][]interface{}

type Candles []Candle

type Candle struct {
	Time  time.Time
	High  float64
	Low   float64
	Open  float64
	Close float64
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
