package user

import (
	"context"
	"fmt"
	"github.com/bitfinexcom/bitfinex-api-go/v2"
	"github.com/vmihailenco/msgpack"
	"log"
	"robot/bfApi"
	"robot/bfSocket"
	"robot/logger"
	"robot/policy"
	"robot/telegramBot"
	"runtime/debug"
	"sync"
	"time"
)

type UserInfo struct {
	TelegramId       int64
	Config           *ConfigManage
	Key              string // bitfinex api key
	Sec              string // bitfinex sec
	Wallet           *Wallet
	API              *bfApi.APIClient
	BFSocket         *bfSocket.Socket
	NotifyChan       chan int
	UpdateWalletChan chan *bitfinex.WalletUpdate

	ctx    context.Context
	cancel context.CancelFunc
}

func (t *UserInfo) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(t)
}

func (t *UserInfo) BinaryUnmarshaler(data []byte) error {
	return msgpack.Unmarshal(data, t)
}

func NewUser(telegramId int64, key, sec string) *UserInfo {
	instance := &UserInfo{
		TelegramId:       telegramId,
		Config:           NewConfig(),
		Key:              key,
		Sec:              sec,
		Wallet:           NewWallet(),
		API:              bfApi.NewAPIClient(),
		BFSocket:         bfSocket.NewSocket(key, sec),
		NotifyChan:       make(chan int),
		UpdateWalletChan: make(chan *bitfinex.WalletUpdate),
	}

	instance.API.RegisterClient(telegramId, key, sec)
	instance.ctx, instance.cancel = context.WithCancel(context.Background())
	go instance.ListenWalletStatus()
	go instance.ListenOnFundingOffer()

	return instance
}

// 監聽 wallet 狀況
func (t *UserInfo) ListenWalletStatus() {
	t.BFSocket.Listen(t.UpdateWalletChan)
	for walletUpdate := range t.UpdateWalletChan {
		t.Wallet.Update(walletUpdate.Balance, walletUpdate.BalanceAvailable)
		if walletUpdate.BalanceAvailable < 50 || t.Config.GetSubmitOffer() == false {
			continue
		}

		// 放貸天數
		day := t.Config.GetDay()
		// 計算放貸利率
		rate := policy.TrackMatchPrice()

		if rate <= t.Config.InValidRate {
			log.Println("計算結果低於: ", rate)
			return
		}

		fixedAmount := t.Config.GetFixedAmount()
		amount := t.Wallet.GetAmount(fixedAmount)
		for amount >= 50 {
			if rate >= t.Config.GetCrazyRate() {
				day = 30
			}

			logger.LOG.Infof("Calculate Rate : %v, sign %v", rate, j)
			err := t.API.SubmitFundingOffer(t.TelegramId, bitfinex.FundingPrefix+"USD", false, amount, rate, int64(day))
			if err != nil {
				logger.LOG.Errorf("UserId [%d] Submit Offer Error: [%v]", t.TelegramId, err)
				break
			}
			rate += t.Config.GetIncreaseRate()
			amount = t.Wallet.GetAmount(fixedAmount)
		}


		//logger.LOG.Infof("Calculate Rate : %v, sign %v", rate, j)

		//if config_manage.Config.GetSubmitOffer() {
		//	for wallet.BalanceAvailable >= 50 {
		//		if rate >= config_manage.Config.GetCrazyRate() {
		//			day = 30
		//		}
		//		fixedAmount := config_manage.Config.GetFixedAmount()
		//		amount := wallet.GetAmount(fixedAmount)
		//		err := bfApi.SubmitFundingOffer(bitfinex.FundingPrefix+"USD", false, amount, rate, int64(day))
		//		if err != nil {
		//			logger.LOG.Errorf("Submit Offer Error: %v", err)
		//			break
		//		}
		//		rate += config_manage.Config.GetIncreaseRate()
		//	}
		//}
	}
}

// 監聽 訂單(未matched)狀況
func (t *UserInfo) ListenOnFundingOffer() {
	wg := sync.WaitGroup{}
	defer func() {
		wg.Done()
		if err := recover(); nil != err {
			bugStack := debug.Stack()
			logger.LOG.Errorf("%v", bugStack)
			logger.LOG.Errorf("on offer error : %v", err)
		}
	}()

	unMatchCount := 0
	var lastUnMatchTimeStamp int64
	wg.Add(1)
loop:
	for {
		select {
		case <-time.After(1 * time.Minute):
			now := time.Now()
			lastFifteenMinute := now.Add(-15 * time.Minute).Unix()
			lastTwoHour := now.Add(-30 * time.Minute).Unix()
			// 每日歸零
			if now.Hour() == 0 && now.Minute() == 0 {
				unMatchCount = 0
			}

			// 權重回歸
			if lastUnMatchTimeStamp < lastTwoHour {
				t.Config.WeightsInit()
			}

			onOfferList := t.API.GetOnOfferList(t.TelegramId)

			if onOfferList != nil {
				for _, offer := range onOfferList {
					if lastFifteenMinute > (offer.MTSCreated / 1000) {
						unMatchCount++

						//if object.UnMatchedCount%3 == 0 {
						//	config_manage.Config.SetWeights("book03", -1)
						//	config_manage.Config.SetWeights("avg100", 1)
						//}
						t.Config.SetWeights("book03", -1)
						t.Config.SetWeights("book01", 1)
						t.Config.SetWeights("avg100", 1)
						t.API.CancelFundingOffer(t.TelegramId, offer.ID)

						lastUnMatchTimeStamp = now.Unix()
						go telegramBot.SendMessage(t.TelegramId, fmt.Sprintf("單號:%d Rate: %f Day: %d ,..超過15分鐘未撮合, 今日已累積未搓合次數:%d", offer.ID, offer.Rate, offer.Period, unMatchCount))
					}
				}
			}
		case <-t.ctx.Done():
			break loop
		}
	}
}

func (t *UserInfo) SubmitOrder() {

}
