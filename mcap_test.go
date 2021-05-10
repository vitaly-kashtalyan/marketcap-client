package mcap

import (
	"testing"
)

var c = NewClient(nil)

func TestCandles_GetCandles(t *testing.T) {
	candles, err := c.GetCandles(CandlesRequest{Symbol: "BTC/USD", Interval: "M1", Limit: 3})
	if err != nil {
		t.Fatalf("error must be empty: %v", err)
	}
	if len(candles) != 3 {
		t.Fatalf("%d != %d", len(candles), 3)
	}
}

func TestErrorCandles_GetCandles(t *testing.T) {
	_, err := c.GetCandles(CandlesRequest{Symbol: "BTC/USD", Interval: "M2", Limit: 3})
	if err == nil {
		t.Fatalf("error must not be empty: %v", err)
	}
	er := err.(*ErrorResponse)
	if er.Code != -1 {
		t.Fatalf("%d != %d", er.Code, -1)
	}
	if er.Msg != "Not supported interval parameter" {
		t.Fatalf("Msg should be: Not supported interval parameter")
	}
}

func TestAssets_GetAssets(t *testing.T) {
	data, err := c.GetAssets()
	if err != nil {
		t.Fatalf("error must be empty: %v", err)
	}
	if len(data) != 10 {
		t.Fatalf("%d != %d", len(data), 10)
	}
}

func TestOrderbook_GetOrderbook(t *testing.T) {
	var d int = 3
	data, err := c.GetOrderbook(OrderbookRequest{Symbol: "BTC/USD", Depth: d})
	if err != nil {
		t.Fatalf("error must be empty: %v", err)
	}
	if len(data.Asks) != d {
		t.Fatalf("%d != %d", len(data.Asks), d)
	}
	if len(data.Bids) != d {
		t.Fatalf("%d != %d", len(data.Bids), d)
	}
}

func TestSummary_GetSummary(t *testing.T) {
	data, err := c.GetSummary()
	if err != nil {
		t.Fatalf("error must be empty: %v", err)
	}
	if len(data.Data) == 0 {
		t.Fatalf("%d more than %d", len(data.Data), 0)
	}
}

func TestTicker_GetTicker(t *testing.T) {
	data, err := c.GetTicker()
	if err != nil {
		t.Fatalf("error must be empty: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("%d more than %d", len(data), 0)
	}
}

func TestTrades_GetTrades(t *testing.T) {
	data, err := c.GetTrades(TradesRequest{Symbol: "BTC/USD"})
	if err != nil {
		t.Fatalf("error must be empty: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("%d more than %d", len(data), 0)
	}
}

func TestErrorResponse_ErrorResponse(t *testing.T) {
	_, err := c.GetCandles(CandlesRequest{Symbol: "BTC/USD", Interval: "M1", Limit: -3})
	if err == nil {
		t.Fatalf("error must not be empty: %v", err)
	}
	expectedMsg := "GET https://marketcap.backend.currency.com/api/v1/candles?interval=M1&limit=-3&symbol=BTC%2FUSD: [400 Bad Request] Invalid limit parameter"
	actualMsg := err.Error()
	if actualMsg != expectedMsg {
		t.Fatalf("error message must be: %v ; but found: %v", expectedMsg, actualMsg)
	}

}
