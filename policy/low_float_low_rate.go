package policy

import (
	"os"
	"sort"
	"strconv"
)

// 低浮動低利率
type LowFloatLowRate struct {
}

func NewLowFloatLowRate() *LowFloatLowRate {
	return &LowFloatLowRate{}
}
func (st *LowFloatLowRate) GetMarketInfo() {

}

func (st *LowFloatLowRate) Execute(marketData *MarketDate) []float64 {

	var rates []float64
	for rate, _ := range marketData.Book {
		rates = append(rates, rate)
	}
	sort.Sort(sort.Float64Slice(rates))

	targetAmount, _ := strconv.ParseFloat(os.Getenv("POLICY_1_TARGET_AMOUNT"), 64)
	increaseAmount, _ := strconv.ParseFloat(os.Getenv("POLICY_1_INCREASE_AMOUNT"), 64)

	calAmount := 0.0
	var rateRrr []float64

	for _, rate := range rates {
		calAmount += marketData.Book[rate]

		if calAmount > targetAmount {
			targetAmount += increaseAmount
			rateRrr = append(rateRrr, rate)
		}
	}

	return rateRrr
}
