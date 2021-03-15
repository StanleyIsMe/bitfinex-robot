package main

import (
	"fmt"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/convert"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/book"
	"github.com/bitfinexcom/bitfinex-api-go/v2/rest"
	"math"
	"net/url"
	"path"
	"sort"

	//"robot/utils"
)

func main() {
	c := rest.NewClientWithURL("https://api.bitfinex.com/v2/")

	req := rest.NewRequestWithMethod(path.Join("book", "fUSD", "P0"), "GET")
	req.Params = make(url.Values)
	req.Params.Add("_full", "1")

	raw, err := c.Request(req)

	if err != nil {
		fmt.Println("err....:", err)
		return
	}

	result, err := book.SnapshotFromRaw("fUSD", "P0", convert.ToInterfaceArray(raw), raw)

	allBook := make(map[float64]float64, 0)
	for _, val := range result.Snapshot {
		if val.Amount < 0 {
			continue
		}
		rate := math.Floor(val.Rate*1000000)/1000000
		allBook[rate] += val.Amount
	}

	var rates []float64
	for rate, _ := range allBook {
		rates = append(rates, rate)
	}

	sort.Sort(sort.Float64Slice(rates))

	calAmount := 0.0
	var rateRrr []float64
	targetAmount := 3000000.0
	for _, rate := range rates {
		calAmount+=allBook[rate]

		if calAmount > targetAmount {
			targetAmount += 2000000
			rateRrr = append(rateRrr, rate)
			fmt.Println(rate, "!!!")

		}
		//utils.PrintWithStruct(val)
	}


	return
}
