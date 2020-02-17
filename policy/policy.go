package policy

import (
	"log"
	"sync"

	"github.com/bitfinexcom/bitfinex-api-go/v2"
	"robot/btApi"
	"robot/config_manage"
)

//type CalculateRate struct {
//	MatchedList []*bitfinex.Trade      // 最近成交價
//	BookListP0  []*bitfinex.BookUpdate //
//	BookListP1  []*bitfinex.BookUpdate
//	BookListP2  []*bitfinex.BookUpdate
//}


type Wallet struct {
	sync.RWMutex

	Balance          float64
	BalanceAvailable float64
	wg               *sync.WaitGroup
}

var myWallet *Wallet

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

func InitPolicy(){
	config := config_manage.NewConfig()
	config.Policy = TrackMatchPrice
}
func TrackBookPrice() float64 {
	log.Println("Use TrackBookPrice Policy")
	config := config_manage.NewConfig()

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
	p0Avg := excueBookAvg(offerListP0, inValidRate)
	p1Avg := excueBookAvg(offerListP1, inValidRate)
	p2Avg := excueBookAvg(offerListP2, inValidRate)
	matchAvg1 := excueMatchedAvg(matchedList[0:100], inValidRate)
	matchAvg2 := excueMatchedAvg(matchedList, inValidRate)
	//allAbg := (p0Avg*5+p1Avg*2+p2Avg*1+matchAvg1*1+matchAvg2*8)/17
	allAbg := (p0Avg + p1Avg + p2Avg + matchAvg1 + matchAvg2) / 5

	bottomRate := config.GetBottomRate()
	if bottomRate == 0 {
		bottomRate = bidListP0[0].Price
	}

	if allAbg < bottomRate {
		return bottomRate
	}

	return allAbg
}

func TrackMatchPrice() float64 {
	log.Println("Use TrackMatchPrice Policy")
	config := config_manage.NewConfig()
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
	p0Avg := excueBookAvg(offerListP0, inValidRate)
	p1Avg := excueBookAvg(offerListP1, inValidRate)
	p2Avg := excueBookAvg(offerListP2, inValidRate)
	matchAvg1 := excueMatchedAvg(matchedList[0:100], inValidRate)
	matchAvg2 := excueMatchedAvg(matchedList, inValidRate)
	//allAbg := (p0Avg*5+p1Avg*2+p2Avg*1+matchAvg1*1+matchAvg2*8)/17
	allAvg := (p0Avg + p1Avg + p2Avg*10 + matchAvg1*1 + matchAvg2*3) / 16

	// 假如沒設定最小利率，則以市場最高出價利率當作最低
	bottomRate := config.GetBottomRate()
	if bottomRate == 0 {
		bottomRate = bidListP0[0].Price
	}

	if bottomRate > allAvg && bottomRate > matchAvg1 {
		return bottomRate
	}

	// 假如算出比近期平均成交利率還低，就以平均成交利率為主
	if allAvg <  matchAvg1 {
		return matchAvg1
	}

	return allAvg
}

func excueBookAvg(list []*bitfinex.BookUpdate, inValidRate float64) (average float64) {
	var count float64
	for _, data := range list {
		if data.Price > inValidRate {
			average += data.Price
			count++
		}
	}
	return average / count
}

func excueMatchedAvg(list []*bitfinex.Trade, inValidRate float64) (average float64) {
	var count float64
	for _, data := range list {
		if data.Price > inValidRate {
			average += data.Price
			count++
		}
	}
	return average / count
}
