package main

import (
	"fmt"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/convert"
	"github.com/bitfinexcom/bitfinex-api-go/v2/rest"
	"robot/model"
)

func main() {
	c := rest.NewClientWithURL("https://api.bitfinex.com/v2/")

	req := rest.NewRequestWithMethod("ticker/fUSD", "GET")
	//req.Params = make(url.Values)
	//req.Params.Add("symbols", strings.Join([]string{"fUSD"}, ","))
	raw, err := c.Request(req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(raw)
	//nraw := raw
	t := &model.TickerCustom{}
	for index, val := range raw {
		switch index {
		case 0:
			t.Frr = convert.F64ValOrZero(val)
			break
		case 1:
			t.Bid = convert.F64ValOrZero(val)
			break
		case 2:
			t.BidPeriod = convert.I64ValOrZero(val)
			break
		case 3:
			t.BidSize = convert.F64ValOrZero(val)
			break
		case 4:
			t.Ask = convert.F64ValOrZero(val)
			break
		case 5:
			t.AskPeriod = convert.I64ValOrZero(val)
			break
		case 6:
			t.AskSize = convert.F64ValOrZero(val)
			break
		case 7:
			t.DailyChange = convert.F64ValOrZero(val)
			break
		case 8:
			t.DailyChangePerc = convert.F64ValOrZero(val)
			break
		case 9:
			t.LastPrice = convert.F64ValOrZero(val)
			break
		case 10:
			t.Volume = convert.F64ValOrZero(val)
			break
		case 11:
			t.High = convert.F64ValOrZero(val)
			break
		case 12:
			t.Low = convert.F64ValOrZero(val)
			break
		case 15:
			t.FrrAmountAvailable = convert.F64ValOrZero(val)
			break
		}
	}

	fmt.Println(t, t.High, t.Low, t.FrrAmountAvailable)

	//tickers := make([]*ticker.Ticker, 0)
	//for _, traw := range raw {
	//	fmt.Println(traw)
	//	//t, err := ticker.FromRestRaw(traw.([]interface{}))
	//
	//
	//}


}

type TickerCustom struct {
	Symbol             string
	Frr                float64
	Bid                float64
	BidPeriod          int64
	BidSize            float64
	Ask                float64
	AskPeriod          int64
	AskSize            float64
	DailyChange        float64
	DailyChangePerc    float64
	LastPrice          float64
	Volume             float64
	High               float64
	Low                float64
	FrrAmountAvailable float64
}
