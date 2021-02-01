package user

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/common"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/wallet"
	"log"
	"robot/bfApi"
	"robot/bfSocket"
	"robot/logger"
	"robot/model"
	"robot/policy"
	"robot/utils"
	"robot/utils/s2c"
	"runtime/debug"
	"sync"
	"time"
)

type UserInfo struct {
	TelegramId       int64               `json:"telegram_id"`
	Config           *ConfigManage       `json:"config"`
	Key              string              `json:"key"` // bitfinex api key
	Sec              string              `json:"sec"` // bitfinex sec
	Wallet           *Wallet             `json:"-"`
	API              *bfApi.APIClient    `json:"-"`
	BFSocket         *bfSocket.Socket    `json:"-"`
	NotifyChan       chan int            `json:"-"`
	UpdateWalletChan chan *wallet.Update `json:"-"`

	CalculateCenter *policy.CalculateCenter `json:"-"`
	ctx             context.Context         `json:"-"`
	cancel          context.CancelFunc      `json:"-"`
}

//func (t *UserInfo) MarshalBinary() ([]byte, error) {
//	return msgpack.Marshal(t)
//}
//
//func (t *UserInfo) BinaryUnmarshaler(data []byte) error {
//	return msgpack.Unmarshal(data, t)
//}

func (t *UserInfo) MarshalBinary() ([]byte, error) {
	return json.Marshal(t)
}

func (t *UserInfo) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	return nil
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
		UpdateWalletChan: make(chan *wallet.Update),
		CalculateCenter:  policy.NewCalculateCenter(),
	}

	instance.API.RegisterClient(telegramId, key, sec)
	instance.ctx, instance.cancel = context.WithCancel(context.Background())
	go instance.ListenWalletStatus()
	go instance.ListenOnFundingOffer()

	return instance
}

func (t *UserInfo) StartActive() {
	t.Wallet = NewWallet()
	t.API = bfApi.NewAPIClient()
	t.BFSocket = bfSocket.NewSocket(t.Key, t.Sec)
	t.NotifyChan = make(chan int)
	t.UpdateWalletChan = make(chan *wallet.Update)
	t.CalculateCenter = policy.NewCalculateCenter()
	t.Config.WeightsInit()
	t.API.RegisterClient(t.TelegramId, t.Key, t.Sec)
	t.ctx, t.cancel = context.WithCancel(context.Background())
	go t.ListenWalletStatus()
	go t.ListenOnFundingOffer()
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
		rate := t.CalculateCenter.CalculateRateByConfig(t.Config.GetWeights())

		bottomRate := t.Config.GetBottomRate()
		if rate <= bottomRate {
			log.Println("計算結果 %v 低於最低利率 %v: ", rate, bottomRate)
			rate = bottomRate
			return
		}

		fixedAmount := t.Config.GetFixedAmount()
		amount := t.Wallet.GetAmount(fixedAmount)
		for amount >= 50 {
			day = t.Config.GetDayByRate(rate)
			//if rate >= t.Config.GetCrazyRate() {
			//	day = 30
			//}

			logger.LOG.Infof("Calculate Rate : %v, wallet %v", rate, walletUpdate)
			err := t.BFSocket.SubmitFundingOffer(common.FundingPrefix+"USD", amount, rate, int64(day))
			//err := t.API.SubmitFundingOffer(t.TelegramId, common.FundingPrefix+"USD", false, amount, rate, int64(day))
			if err != nil {
				logger.LOG.Errorf("UserId [%d] Submit Offer Error: [%v]", t.TelegramId, err)
				break
			}
			rate += t.Config.GetIncreaseRate()
			amount = t.Wallet.GetAmount(fixedAmount)
		}
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
			cancelTime := t.Config.GetAutoCancelTime()
			lastFifteenMinute := now.Add(time.Duration(-cancelTime) * time.Minute).Unix()
			lastTwoHour := now.Add(-30 * time.Minute).Unix()
			// 每日歸零
			if now.Hour() == 0 && now.Minute() == 0 {
				unMatchCount = 0
			}

			// 權重回歸
			if lastUnMatchTimeStamp < lastTwoHour {
				t.Config.WeightsInit()
			}

			// sync wallet info
			t.BFSocket.CalWalletUpdate()
			if ws := t.API.Wallets(t.TelegramId); ws != nil {
				for _, wallets := range ws.Snapshot {
					if wallets.Type == "funding" && wallets.Currency == "USD"{
						utils.PrintWithStruct(wallets)
						newWalletUpdate := &wallet.Update{
							Balance:          wallets.Balance,
							BalanceAvailable: wallets.BalanceAvailable,
						}
						t.UpdateWalletChan <- newWalletUpdate
					}
				}
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
						//t.API.CancelFundingOffer(t.TelegramId, offer.ID)
						t.BFSocket.CancelFundingOffer(offer.ID)
						lastUnMatchTimeStamp = now.Unix()
						go s2c.SendMessage(t.TelegramId, fmt.Sprintf("單號:%d Rate: %f Day: %d ,..超過%d分鐘未撮合, 今日已累積未搓合次數:%d", offer.ID, offer.Rate, offer.Period, cancelTime, unMatchCount))
					}
				}
			}
		case <-t.ctx.Done():
			break loop
		}
	}
}

func (t *UserInfo) GetFundingRate() float64 {
	return t.CalculateCenter.CalculateRateByConfig(t.Config.GetWeights())
}

func (t *UserInfo) GetInterest() *model.DailyInterestReport {
	report := &model.DailyInterestReport{}

	end := time.Now().UnixNano() / int64(time.Millisecond)
	list := t.API.GetLedgers(t.TelegramId, end)
	count := 0
	for len(list) > 0 {
		for _, data := range list {

			if data.Description == "Margin Funding Payment on wallet funding" {
				count++

				// 第一筆為總金額
				if count == 1 {
					report.Balance = data.Balance
				}

				report.TotalInterest += data.Amount

				if count > 10 {
					continue
				}
				earnInfo := map[string]interface{}{}
				dateTime := time.Unix(data.MTS/1000, 0).Format("2006-01-02 15:04:05")
				earnInfo["Date"] = dateTime
				earnInfo["Interest"] = data.Amount
				report.InterestList = append(report.InterestList, earnInfo)
			}
			end = data.MTS
		}

		list = t.API.GetLedgers(t.TelegramId, end-1)
	}

	return report
}
