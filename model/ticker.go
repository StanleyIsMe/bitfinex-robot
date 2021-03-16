package model

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

