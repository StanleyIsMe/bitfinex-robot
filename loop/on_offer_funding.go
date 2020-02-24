package loop

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"robot/bfApi"
	"robot/config_manage"
	"robot/logger"
	"robot/telegramBot"
)

type LoopOnOffer struct {
	ctx            context.Context
	cancel         context.CancelFunc
	wg             *sync.WaitGroup
	last           int64
	UnMatchedCount int
}

func NewLoopOnOffer() *LoopOnOffer {
	object := &LoopOnOffer{
		wg: &sync.WaitGroup{},
	}
	object.ctx, object.cancel = context.WithCancel(context.Background())
	go object.loop()
	return object
}

func (object *LoopOnOffer) loop() {

	defer func() {
		if err := recover(); nil != err {
			bugStack := debug.Stack()
			logger.LOG.Errorf("%v", bugStack)
			logger.LOG.Errorf("on offer error : %v", err)
			object.wg.Done()
		}
	}()

	object.wg.Add(1)
loop:
	for {
		select {
		case <-time.After(1 * time.Minute):
			now := time.Now()
			lastFifteenMinute := now.Add(-15 * time.Minute).Unix()
			lastTwoHour := now.Add(-30 * time.Minute).Unix()
			// 每日歸零
			if now.Hour() == 0 && now.Minute() == 0 {
				object.UnMatchedCount = 0
			}

			// 權重回歸
			if object.last < lastTwoHour {
				config_manage.Config.WeightsInit()
			}

			onOfferList := bfApi.GetOnOfferList()

			if onOfferList != nil {
				for _, offer := range onOfferList {
					if lastFifteenMinute > (offer.MTSCreated / 1000) {
						object.UnMatchedCount++

						//if object.UnMatchedCount%3 == 0 {
						//	config_manage.Config.SetWeights("book03", -1)
						//	config_manage.Config.SetWeights("avg100", 1)
						//}
						config_manage.Config.SetWeights("book03", -1)
						config_manage.Config.SetWeights("book01", 1)
						config_manage.Config.SetWeights("avg100", 1)
						bfApi.CancelFundingOffer(offer.ID)

						object.last = now.Unix()
						go telegramBot.SendMessage(config_manage.Config.TelegramId, fmt.Sprintf("單號:%d Rate: %f Day: %d ,..超過15分鐘未撮合, 今日已累積未搓合次數:%d", offer.ID, offer.Rate, offer.Period, object.UnMatchedCount))
					}
				}
			}
		case <-object.ctx.Done():
			break loop
		}
	}
	object.wg.Done()
}

func (object *LoopOnOffer) ShutDown() {
	object.cancel()
	object.wg.Wait()
}
