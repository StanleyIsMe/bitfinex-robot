package main

import (
	"fmt"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/common"
	"github.com/bitfinexcom/bitfinex-api-go/v2/rest"
	"robot/utils"
	"sort"
)

func main() {
	c := rest.NewClientWithURL("https://api.bitfinex.com/v2/")

	result, err := c.Book.All(common.FundingPrefix+"USD", "P4", 100)
	if err != nil {
		fmt.Println(err)
	}
	total := make(map[float64]float64, 0)

	for _, val := range result.Snapshot[100:] {
		total[val.Rate] += val.Amount
	}
	utils.PrintWithStruct(total)
	var rates []float64
	for rate, _ := range total {
		rates = append(rates, rate)
	}
	//sort.Float64s(tt)
	sort.Sort(sort.Reverse(sort.Float64Slice(rates)))

	//for _, rate := range tt {
	//	utils.PrintWithStruct(rate, total[rate])
	//}

	var totalRate, count float64
	for _, rate := range rates {
		fmt.Println(rate,"!")
		totalRate += rate
		count++
		if rate > 0.07 {
			continue
		}
		if total[rate] <= 6000000 {
			fmt.Println((rate+0.0014)/2, "here")
			return
		}
	}
	fmt.Println(totalRate/count, "here3")
}
