package policy

import "sort"

// 高浮動高利率
type HighFloatHighRate struct {
}

func NewHighFloatHighRate() *HighFloatHighRate {
	return &HighFloatHighRate{}
}
func (st *HighFloatHighRate) GetMarketInfo() {

}

func (st *HighFloatHighRate) Execute(marketData *MarketDate) float64 {
	var rates []float64
	for rate, _ := range marketData.Book["P4"] {
		rates = append(rates, rate)
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(rates)))

	var totalRate, count float64
	for _, rate := range rates {
		totalRate += rate
		count++
		if rate >= marketData.DailyHighRate {
			continue
		}
		if marketData.Book["P4"][rate] <= 6000000 {
			return rate
		}
		return rate
	}
	return totalRate/count
}
