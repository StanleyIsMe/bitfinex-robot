package policy

import "sort"

// 低浮動低利率
type LowFloatLowRate struct {
}

func NewLowFloatLowRate() *LowFloatLowRate {
	return &LowFloatLowRate{}
}
func (st *LowFloatLowRate) GetMarketInfo() {

}

func (st *LowFloatLowRate) Execute(marketData *MarketDate) float64 {

	var rates []float64
	for rate, _ := range marketData.Book["P3"] {
		rates = append(rates, rate)
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(rates)))
	var totalRate, count float64
	for _, rate := range rates {
		if rate > marketData.FRR && marketData.FRRVolume > 30000000 {
			continue
		}

		if count >= 5 {
			break
		}
		totalRate += rate
		count++
	}
	//for rate, amount := range marketData.Book["P3"] {
	//	if rate > marketData.FRR && marketData.FRRVolume > 30000000 {
	//		continue
	//	}
	//
	//	//  30w~300w
	//	if amount > 300000 && amount < 3000000 {
	//		totalRate += rate
	//		count++
	//	}
	//}

	if count == 0 {
		return marketData.FRR
	}
	return totalRate / count
}
