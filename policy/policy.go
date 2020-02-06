package policy

import (
	"os"
	"strconv"
	"sync"

	"github.com/bitfinexcom/bitfinex-api-go/v2"
	"robot/btApi"
)

type CalculateRate struct {
	MatchedList []*bitfinex.Trade      // 最近成交價
	BookListP0  []*bitfinex.BookUpdate //
	BookListP1  []*bitfinex.BookUpdate
	BookListP2  []*bitfinex.BookUpdate
}

type RateControl struct {
	TopRate      float64
	BottomRate   float64
	FixedAmount  float64
	Day          int
	CrazyRate    float64
	NormalRate   float64
	IncreaseRate float64
}

type FundingInfo struct {
	Loaned  []int
	Lending []int
}

type Wallet struct {
	sync.RWMutex

	Balance          float64
	BalanceAvailable float64
	wg               *sync.WaitGroup
}

var MyRateController *RateControl
var myWallet *Wallet
var MyOnFunding []int

func NewWallet() *Wallet {
	var once sync.Once

	if myWallet == nil {
		once.Do(func() {
			myWallet = &Wallet{
				Balance:          0,
				BalanceAvailable: 0,
				wg:               &sync.WaitGroup{},
			}
		})
	}

	return myWallet
}

func (object *Wallet) Update(balance, balanceAvailable float64) {
	object.Lock()
	object.Balance = balance
	object.BalanceAvailable = balanceAvailable
	object.Unlock()
}

func (object *Wallet) GetAmount(basicAmount float64) float64 {
	minimumAmount := 50.0
	object.Lock()
	defer object.Unlock()

	if ((object.BalanceAvailable-basicAmount) < minimumAmount ) || (object.BalanceAvailable <= basicAmount) {
		temp := object.BalanceAvailable
		object.BalanceAvailable = 0
		return temp
	}
	object.BalanceAvailable -= basicAmount
	return basicAmount
}

func PolicyInit() {
	bottomRate, _ := strconv.ParseFloat(os.Getenv("FUNDING_BOTTOM_RATE"), 64)
	topRate, _ := strconv.ParseFloat(os.Getenv("FUNDING_TOP_RATE"), 64)
	crazyRate, _ := strconv.ParseFloat(os.Getenv("FUNDING_CRAZY_RATE"), 64)
	normalRate, _ := strconv.ParseFloat(os.Getenv("FUNDING_NORMAL_RATE"), 64)
	fixedAmount, _ := strconv.ParseFloat(os.Getenv("FUNDING_FIXED_AMOUNT"), 64)
	increaseRate, _ := strconv.ParseFloat(os.Getenv("FUNDING_INCREASE_RATE"), 64)

	MyRateController = &RateControl{
		TopRate:      topRate,
		BottomRate:   bottomRate,
		FixedAmount:  fixedAmount,
		Day:          2,
		CrazyRate:    crazyRate,
		NormalRate:   normalRate,
		IncreaseRate: increaseRate,
	}

}

func AllocationFunds() {

}

func TrackBookPrice() float64 {
	// 無效值先隨意暫定
	inValidRate := 0.0003

	bidListP0, offerListP0, err0 := bfApi.GetBook(bitfinex.Precision0)
	_, offerListP1, err1 := bfApi.GetBook(bitfinex.Precision1)
	_, offerListP2, err2 := bfApi.GetBook(bitfinex.Precision2)
	matchedList, err := bfApi.GetMatched(10000)
	if err != nil || err0 != nil || err1 != nil || err2 != nil {
		return 0
	}

	// 算市場平均價
	p0Avg := bookAvg(offerListP0, inValidRate)
	p1Avg := bookAvg(offerListP1, inValidRate)
	p2Avg := bookAvg(offerListP2, inValidRate)
	matchAvg1 := matchedAvg(matchedList[0:100], inValidRate)
	matchAvg2 := matchedAvg(matchedList, inValidRate)
	//allAbg := (p0Avg*5+p1Avg*2+p2Avg*1+matchAvg1*1+matchAvg2*8)/17
	allAbg := (p0Avg + p1Avg + p2Avg + matchAvg1 + matchAvg2) / 5

	bottomRate := MyRateController.BottomRate
	if bottomRate == 0 {
		bottomRate = bidListP0[0].Price
	}

	if allAbg < bottomRate {
		return bottomRate
	}

	return allAbg
}

func TrackMatchPrice() float64 {
	// 無效值先隨意暫定
	inValidRate := 0.0003

	bidListP0, offerListP0, err0 := bfApi.GetBook(bitfinex.Precision0)
	_, offerListP1, err1 := bfApi.GetBook(bitfinex.Precision1)
	_, offerListP2, err2 := bfApi.GetBook(bitfinex.Precision2)
	matchedList, err := bfApi.GetMatched(10000)
	if err != nil || err0 != nil || err1 != nil || err2 != nil {
		return 0
	}

	// 算市場平均價
	p0Avg := bookAvg(offerListP0, inValidRate)
	p1Avg := bookAvg(offerListP1, inValidRate)
	p2Avg := bookAvg(offerListP2, inValidRate)
	matchAvg1 := matchedAvg(matchedList[0:100], inValidRate)
	matchAvg2 := matchedAvg(matchedList, inValidRate)
	//allAbg := (p0Avg*5+p1Avg*2+p2Avg*1+matchAvg1*1+matchAvg2*8)/17
	allAbg := (p0Avg + p1Avg + p2Avg*10 + matchAvg1*1 + matchAvg2*3) / 16

	bottomRate := MyRateController.BottomRate
	if bottomRate == 0 {
		bottomRate = bidListP0[0].Price
	}

	if bottomRate > allAbg && bottomRate > matchAvg1 {
		return bottomRate
	}

	if allAbg <  matchAvg1 {
		return matchAvg1
	}

	return allAbg
}

func bookAvg(list []*bitfinex.BookUpdate, inValidRate float64) (average float64) {
	var count float64
	for _, data := range list {
		if data.Price > inValidRate {
			average += data.Price
			count++
		}
	}
	return average / count
}

func matchedAvg(list []*bitfinex.Trade, inValidRate float64) (average float64) {
	var count float64
	for _, data := range list {
		if data.Price > inValidRate {
			average += data.Price
			count++
		}
	}
	return average / count
}
