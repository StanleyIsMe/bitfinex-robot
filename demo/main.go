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

	result, err := c.Book.All(common.FundingPrefix+"USD", "P3", 100)
	if err != nil {
		fmt.Println(err)
	}
	total := make(map[float64]float64, 0)

	for _, val := range result.Snapshot[100:] {
		total[val.Rate] += val.Amount
	}
	utils.PrintWithStruct(total)
	var tt []float64
	for index, _ := range total {
		tt = append(tt, index)
	}
	//sort.Float64s(tt)
	sort.Sort(sort.Reverse(sort.Float64Slice(tt)))

	for _, rate := range tt {
		utils.PrintWithStruct(rate, total[rate])
	}

}

// 低浮動低利率 0.0002~0.0005  (FRR 0.0005, volum 3000w up)
// 低浮動高利率 0.0005~0.001 (FRR 0.001 volum 3000w up)
// 高浮動中低利率 0.0003~0.0005 (FRR 0.00045 volum 500w左右
// 高浮動高利率 0.0006~0.009 (FRR 0.0007 volum 500w 左右)
