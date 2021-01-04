package policy

import (
	"log"
	"sync"

	"github.com/bitfinexcom/bitfinex-api-go/v2"
	"robot/bfApi"
	"robot/config_manage"
)

type Wallet struct {
	sync.RWMutex

	Balance          float64
	BalanceAvailable float64
	wg               *sync.WaitGroup
}

var myWallet *Wallet
var walletSingleton sync.Once

func NewWallet() *Wallet {
	walletSingleton.Do(func() {
		myWallet = &Wallet{
			Balance:          0,
			BalanceAvailable: 0,
			wg:               &sync.WaitGroup{},
		}
	})

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

	if ((object.BalanceAvailable - basicAmount) < minimumAmount) || (object.BalanceAvailable <= basicAmount) {
		temp := object.BalanceAvailable
		object.BalanceAvailable = 0
		return temp
	}
	object.BalanceAvailable -= basicAmount
	return basicAmount
}

func InitPolicy() {
	//config := config_manage.NewConfig()
	//config.Policy = TrackMatchPrice
}

func TrackMatchPrice() float64 {
	log.Println("Use TrackMatchPrice Policy")


	bidListP0, offerListP0, err0 := bfApi.GetBook(bitfinex.Precision0)
	_, offerListP1, err1 := bfApi.GetBook(bitfinex.Precision1)
	_, offerListP2, err2 := bfApi.GetBook(bitfinex.Precision2)
	matchedList, err := bfApi.GetMatched(10000)
	if err != nil || err0 != nil || err1 != nil || err2 != nil {
		return 0
	}

	// 算市場平均價
	p0Avg := excueBookAvg(offerListP0, config_manage.Config.GetInValidRate())
	p1Avg := excueBookAvg(offerListP1, config_manage.Config.GetInValidRate())
	p2Avg := excueBookAvg(offerListP2, config_manage.Config.GetInValidRate())
	matchAvg1 := excueMatchedAvg(matchedList[0:100], config_manage.Config.GetInValidRate())
	matchAvg2 := excueMatchedAvg(matchedList, config_manage.Config.GetInValidRate())

	weights := config_manage.Config.GetWeights()
	var allAvg float64
	total := 0
	for key, weight := range weights {
		switch key {
		case "book01":
			allAvg += p0Avg * float64(weight)
			total += weight
			break
		case "book02":
			allAvg += p1Avg * float64(weight)
			total += weight
			break
		case "book03":
			allAvg += p2Avg * float64(weight)
			total += weight
			break
		case "avg100":
			allAvg += matchAvg1 * float64(weight)
			total += weight
			break
		case "avg10000":
			allAvg += matchAvg2 * float64(weight)
			total += weight
			break

		}
	}

	allAvg = allAvg / float64(total)


	// 假如沒設定最小利率，則以市場最高出價利率當作最低
	bottomRate := config_manage.Config.GetBottomRate()
	if bottomRate == 0 {
		bottomRate = bidListP0[0].Price
	}

	if bottomRate > allAvg && bottomRate > matchAvg1 {
		return bottomRate
	}

	// 假如算出比近期平均成交利率還低，就以平均成交利率為主
	if allAvg < matchAvg1 {
		return matchAvg1
	}

	return allAvg
}

func TrackMatchPrice3() float64 {
	log.Println("Use TrackMatchPrice Policy")

	// 無效值先隨意暫定
	inValidRate := config_manage.Config.GetInValidRate()

	bidListP0, offerListP0, err0 := bfApi.GetBook(bitfinex.Precision0)
	_, offerListP1, err1 := bfApi.GetBook(bitfinex.Precision1)
	_, offerListP2, err2 := bfApi.GetBook(bitfinex.Precision2)
	_, offerListP3, err3 := bfApi.GetBook(bitfinex.Precision3)
	matchedList, err := bfApi.GetMatched(10000)
	if err != nil || err0 != nil || err1 != nil || err2 != nil || err3 != nil{
		return 0
	}

	// 算市場平均價
	p0Avg := excueBookAvg(offerListP0, inValidRate)
	p1Avg := excueBookAvg(offerListP1, inValidRate)
	p2Avg := excueBookAvg(offerListP2, inValidRate)
	p3Avg := excueBookAvg(offerListP3, inValidRate)
	matchAvg1 := excueMatchedAvg(matchedList[0:100], inValidRate)
	matchAvg2 := excueMatchedAvg(matchedList, inValidRate)

	weights := config_manage.Config.GetWeights()
	var allAvg float64
	total := 0
	for key, weight := range weights {
		switch key {
		case "book01":
			allAvg += p0Avg * float64(weight)
			total += weight
			break
		case "book02":
			allAvg += p1Avg * float64(weight)
			total += weight
			break
		case "book03":
			allAvg += p2Avg * float64(weight)
			total += weight
			break
		case "book04":
			allAvg += p3Avg * float64(weight)
			total += weight
			break
		case "avg100":
			allAvg += matchAvg1 * float64(weight)
			total += weight
			break
		case "avg10000":
			allAvg += matchAvg2 * float64(weight)
			total += weight
			break

		}
	}

	allAvg = allAvg / float64(total)


	// 假如沒設定最小利率，則以市場最高出價利率當作最低
	bottomRate := config_manage.Config.GetBottomRate()
	if bottomRate == 0 {
		bottomRate = bidListP0[0].Price
	}

	if bottomRate > allAvg && bottomRate > matchAvg1 {
		return bottomRate
	}

	// 假如算出比近期平均成交利率還低，就以平均成交利率為主
	if allAvg < matchAvg1 {
		return matchAvg1
	}

	return allAvg
}

func TrackMatchPrice2() float64 {

	// 無效值先隨意暫定
	inValidRate := 0.00015

	matchedList, err := bfApi.GetMatched(1000)
	if err != nil {
		return 0
	}

	// 算市場平均價
	matchAvg1 := excueMatchedAvg(matchedList[0:100], inValidRate)

	// 假如沒設定最小利率，則以市場最高出價利率當作最低
	bottomRate := config_manage.Config.GetBottomRate()

	if bottomRate > matchAvg1 {
		return bottomRate
	}

	return matchAvg1
}

func excueBookAvg(list []*bitfinex.BookUpdate, inValidRate float64) (average float64) {
	var count float64
	for _, data := range list {
		if data.Price > inValidRate {
			average += data.Price
			count++
		}
	}

	if count == 0 {
		return 0
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
