package policy

import (
	"context"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/book"
	bfcommon "github.com/bitfinexcom/bitfinex-api-go/pkg/models/common"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/trade"
	"os"
	"robot/bfApi"
	"robot/common"
	"robot/logger"
	"strconv"
	"sync"
	"time"
)

var centerOnce sync.Once
var centerInstance *CalculateCenter

type CalculateCenter struct {
	HistoryRate float64
	MarketData  *MarketDate
	apiClient   *bfApi.APIClient
	sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
}

func NewCalculateCenter() *CalculateCenter {
	centerOnce.Do(func() {
		centerInstance = &CalculateCenter{
			HistoryRate: 0,
			apiClient:   bfApi.NewAPIClient(),
		}

		centerInstance.ctx, centerInstance.cancel = context.WithCancel(context.Background())
		go centerInstance.LoopClearMarketRate()

	})
	return centerInstance
}

type MarketDate struct {
	BidMaxRate    float64 // 最高承租利率
	DailyHighRate float64 // 每日最高利率
	FRR           float64
	FRRVolume     float64
	Book          map[float64]float64 //book  map[P0~P4]map[rate]Amount
	AvgPriceMap   map[string]float64
}

func (center *CalculateCenter) GetMarketPrice() *MarketDate {
	center.Lock()
	defer center.Unlock()

	if center.MarketData != nil {
		return center.MarketData
	}

	//bidListP0, offerListP0, err0 := center.apiClient.GetBook(bfcommon.Precision0)
	//_, offerListP1, err1 := center.apiClient.GetBook(bfcommon.Precision1)
	//_, offerListP2, err2 := center.apiClient.GetBook(bfcommon.Precision2)
	//_, offerListP3, err3 := center.apiClient.GetBook(bfcommon.Precision3)
	//_, offerListP4, err4 := center.apiClient.GetBook("P4")
	offerList, err := center.apiClient.GetAllBook()
	tick := center.apiClient.GetTicker(bfcommon.FundingPrefix + "USD")
	matchedList, err := center.apiClient.GetMatched(10000)
	//if err != nil || err0 != nil || err1 != nil || err2 != nil || err3 != nil || err4 != nil || len(bidListP0) == 0 || tick == nil {
	//	logger.LOG.Error("計算市場利率發生錯誤:", err, err0, err1, err2, err3, err4)
	//	return nil
	//}

	if err != nil  || tick == nil {
		logger.LOG.Error("計算市場利率發生錯誤:", err)
		return nil
	}

	invalidRate, _ := strconv.ParseFloat(os.Getenv("INVALID_RATE"), 64)
	center.MarketData = &MarketDate{
		DailyHighRate: tick.High,
		FRR:           tick.Frr,
		FRRVolume:     tick.FrrAmountAvailable,
		Book: offerList,
		AvgPriceMap: map[string]float64{
			"avg100":   excueMatchedAvg(matchedList[0:100], invalidRate),
			"avg10000": excueMatchedAvg(matchedList, invalidRate),
		},
	}

	return center.MarketData
}

func (center *CalculateCenter) LoopClearMarketRate() {
loop:
	for {
		select {
		case <-time.After(1 * time.Minute):
			center.Lock()
			center.MarketData = nil
			center.Unlock()
		case <-center.ctx.Done():
			break loop
		}
	}
}

//func (center *CalculateCenter) CalculateRateByConfig(weights map[string]int) float64 {
//	marketRateMap := center.GetMarketPrice()
//
//	if marketRateMap == nil {
//		return 0
//	}
//
//	var allAvg float64
//	total := 0
//
//	for key, weight := range weights {
//		if rate, ok := marketRateMap[key]; ok {
//			allAvg += rate * float64(weight)
//			total += weight
//		}
//	}
//
//	allAvg = allAvg / float64(total)
//
//	// 假如沒設定最小利率，則以市場最高出價利率當作最低
//	//bottomRate := config_manage.Config.GetBottomRate()
//	//if bottomRate == 0 {
//	//	bottomRate = bidListP0[0].Price
//	//}
//
//	//if bottomRate > allAvg && bottomRate > matchAvg1 {
//	//	return bottomRate
//	//}
//
//	// 假如算出比近期平均成交利率還低，就以平均成交利率為主
//	if matchAvg100, ok := marketRateMap["avg100"]; ok {
//		if allAvg < matchAvg100 {
//			return matchAvg100
//		}
//	}
//
//	return allAvg
//}

func (center *CalculateCenter) CalculateRateByStrategy(strategy common.StrategyType) []float64 {
	marketData := center.GetMarketPrice()

	if marketData == nil {
		return nil
	}

	var executeResult []float64
	switch strategy {
	case common.LowFloatLowRate:
		obj := NewLowFloatLowRate()
		executeResult = obj.Execute(marketData)
		break
	case common.ExpectHighRate:
		obj := NewExpectHighRate()
		executeResult = obj.Execute(marketData)
		break
	case common.HighFloatHighRate:
		obj := NewHighFloatHighRate()
		executeResult = obj.Execute(marketData)
		break
	default:
		executeResult = []float64{}
	}

	// 假如算出比近期平均成交利率還低，就以平均成交利率為主
	if matchAvg100, ok := marketData.AvgPriceMap["avg100"]; ok {

		if len(executeResult) == 0 {
			return []float64{matchAvg100}
		}

		if executeResult[0] < matchAvg100 {
			return []float64{matchAvg100}
		}
	}

	return executeResult
}


func (center *CalculateCenter) ArrangeBookData(list []*book.Book) map[float64]float64 {
	if len(list) == 0 {
		return nil
	}

	result := make(map[float64]float64, 0)
	for _, val := range list {
		result[val.Rate] += val.Amount
	}
	return result
}


func excueBookAvg(list []*book.Book, inValidRate float64) (average float64) {
	var count float64
	for _, data := range list {
		if data.Rate > inValidRate {
			average += data.Rate
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return average / count
}

func excueMatchedAvg(list []*trade.Trade, inValidRate float64) (average float64) {
	var count float64
	for _, data := range list {
		if data.Rate > inValidRate {
			average += data.Rate
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return average / count
}

//func MultiTrack() {
//	bidListP0, offerListP0, err0 := bfApi.GetBook(bitfinex.Precision0)
//
//	utils.InstanceRoutinePool.PostTask(func(params []interface{}) interface{} {
//		return nil
//	}, bitfinex.Precision1)
//	_, offerListP1, err1 := bfApi.GetBook(bitfinex.Precision1)
//	_, offerListP2, err2 := bfApi.GetBook(bitfinex.Precision2)
//	matchedList, err := bfApi.GetMatched(10000)
//	if err != nil || err0 != nil || err1 != nil || err2 != nil {
//		return 0
//	}
//
//	// 算市場平均價
//	p0Avg := excueBookAvg(offerListP0, config_manage.Config.GetInValidRate())
//	p1Avg := excueBookAvg(offerListP1, config_manage.Config.GetInValidRate())
//	p2Avg := excueBookAvg(offerListP2, config_manage.Config.GetInValidRate())
//	matchAvg1 := excueMatchedAvg(matchedList[0:100], config_manage.Config.GetInValidRate())
//	matchAvg2 := excueMatchedAvg(matchedList, config_manage.Config.GetInValidRate())
//
//	weights := config_manage.Config.GetWeights()
//	var allAvg float64
//	total := 0
//	for key, weight := range weights {
//		switch key {
//		case "book01":
//			allAvg += p0Avg * float64(weight)
//			total += weight
//			break
//		case "book02":
//			allAvg += p1Avg * float64(weight)
//			total += weight
//			break
//		case "book03":
//			allAvg += p2Avg * float64(weight)
//			total += weight
//			break
//		case "avg100":
//			allAvg += matchAvg1 * float64(weight)
//			total += weight
//			break
//		case "avg10000":
//			allAvg += matchAvg2 * float64(weight)
//			total += weight
//			break
//
//		}
//	}
//
//	allAvg = allAvg / float64(total)
//
//
//	// 假如沒設定最小利率，則以市場最高出價利率當作最低
//	bottomRate := config_manage.Config.GetBottomRate()
//	if bottomRate == 0 {
//		bottomRate = bidListP0[0].Price
//	}
//
//	if bottomRate > allAvg && bottomRate > matchAvg1 {
//		return bottomRate
//	}
//
//	// 假如算出比近期平均成交利率還低，就以平均成交利率為主
//	if allAvg < matchAvg1 {
//		return matchAvg1
//	}
//
//	return allAvg
//}
