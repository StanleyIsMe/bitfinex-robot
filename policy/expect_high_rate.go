package policy

// 低利率時，可能出現高利率突破
type ExpectHighRate struct {
}

func NewExpectHighRate() *ExpectHighRate {
	return &ExpectHighRate{}
}
func (st *ExpectHighRate) GetMarketInfo() {

}

func (st *ExpectHighRate) Execute(marketData *MarketDate) []float64 {
	return nil
	//var totalRate, count float64
	//for rate, _ := range marketData.Book["P4"] {
	//	if rate <= marketData.FRR {
	//		continue
	//	}
	//
	//	totalRate += rate
	//	count++
	//	//if amount > 300000 && amount < 3000000 {
	//	//	p3TotalRate += rate
	//	//	p3Count++
	//	//}
	//}
	//
	//if count == 0 {
	//	return marketData.FRR
	//}
	//return totalRate / count
}
