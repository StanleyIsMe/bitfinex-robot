package policy

type Strategy interface {
	GetMarketInfo()
	Execute(marketData *MarketDate) float64
}

