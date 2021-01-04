package bfApi

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/bitfinexcom/bitfinex-api-go/v2"
	"robot/logger"
	"robot/utils"

	"github.com/bitfinexcom/bitfinex-api-go/v2/rest"
)

const RateLimit int8 = 15

type APIClient struct {
	sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	rateCount    int8
	ClientList   map[int64]*rest.Client
	PublicClient *rest.Client
}

var APIOnce sync.Once
var APIClientInstance *APIClient

func NewAPIClient() *APIClient {
	//url := os.Getenv("BFX_API_URI")
	//return &APIClient{
	//	Client: rest.NewClientWithURL(url).Credentials(key, secret),
	//}
	APIOnce.Do(func() {
		APIClientInstance = &APIClient{
			rateCount:  10,
			ClientList: make(map[int64]*rest.Client, 0),
		}

		APIClientInstance.ctx, APIClientInstance.cancel = context.WithCancel(context.Background())
	})
	return APIClientInstance
}

func (api *APIClient) LoopCalculateRateLimit() {
loop:
	for {
		select {
		case <-time.After(1 * time.Minute):
			api.Lock()
			logger.LOG.Info("API Rate 歸零")
			api.rateCount = 0
			api.Unlock()
		case <-api.ctx.Done():
			break loop
		}
	}
}

func (api *APIClient) CheckRateCount() error {
	api.Lock()
	defer api.Unlock()

	if api.rateCount >= RateLimit {
		return errors.New("Reached API Rate limit")
	}
	api.rateCount++
	return nil
}

func (api *APIClient) GetClientByUserId(userId int64) *rest.Client {
	if userClient, ok := api.ClientList[userId]; ok {
		return userClient
	}
	return nil
}

func (api *APIClient) RegisterClient(userId int64, key, secret string) bool {
	api.Lock()
	defer api.Unlock()

	url := os.Getenv("BFX_API_URI")
	if _, ok := api.ClientList[userId]; !ok {
		tempClient := rest.NewClientWithURL(url).Credentials(key, secret)
		if _, err := tempClient.Wallet.Wallet(); err == nil {
			api.ClientList[userId] = tempClient
			return true
		}
		return false
	}
	return true
}

// 每日funding offer 利息獲得及總資產
func (api *APIClient) GetLedgers(userId, end int64) []*bitfinex.Ledger {
	if api.CheckRateCount() != nil {
		return nil
	}

	if client := api.GetClientByUserId(userId); client != nil {
		result, err := client.Ledgers.Ledgers("USD", 0, end, 500)
		if err != nil {
			logger.LOG.Errorf("getting Ledgers: %s", err)
			return nil
		}

		return result.Snapshot
	}
	return nil
}

func (api *APIClient) GetBook(precision bitfinex.BookPrecision) (bid []*bitfinex.BookUpdate, offer []*bitfinex.BookUpdate, err error) {
	if api.CheckRateCount() != nil {
		return
	}

	book, err := api.PublicClient.Book.All(bitfinex.FundingPrefix+"USD", precision, 100)

	if err != nil {
		logger.LOG.Errorf("Get book list: %s", err)
		return
	}

	return book.Snapshot[0:100], book.Snapshot[100:], nil
}

func (api *APIClient) GetMatched(limit int) ([]*bitfinex.Trade, error) {
	if api.CheckRateCount() != nil {
		return nil, nil
	}

	fiveMin, _ := time.ParseDuration("-2h")

	now := time.Now()
	start := bitfinex.Mts(now.Add(fiveMin).UnixNano() / int64(time.Millisecond))
	end := bitfinex.Mts(now.UnixNano() / int64(time.Millisecond))

	matchedList, err := api.PublicClient.Trades.PublicHistoryWithQuery(bitfinex.FundingPrefix+"USD", start, end, bitfinex.QueryLimit(limit), bitfinex.NewestFirst)

	if err != nil {
		logger.LOG.Errorf("Get Matched list: %v", err)
		return nil, err
	}

	return matchedList.Snapshot, nil
}

func (api *APIClient) GetOnOfferList(userId int64) []*bitfinex.Offer {
	if api.CheckRateCount() != nil {
		return nil
	}

	client := api.GetClientByUserId(userId)
	if client == nil {
		logger.LOG.Errorf("UserId %d Not Found", userId)
		return nil
	}

	snap, err := client.Funding.Offers("fUSD")
	if err != nil {
		logger.LOG.Errorf("GetOnOfferList error : %v", err)
		return nil
	}

	if snap != nil {
		return snap.Snapshot
	}
	return nil
}

func (api *APIClient) SubmitFundingOffer(userId int64, symbol string, ffr bool, amount float64, rate float64, day int64) error {
	if api.CheckRateCount() != nil {
		return nil
	}

	client := api.GetClientByUserId(userId)
	if client == nil {
		return errors.New(fmt.Sprintf("UserId %d Not Found", userId))
	}

	fundingType := "LIMIT"
	if ffr {
		fundingType = "FRRDELTAVAR"
	}

	fo, err := client.Funding.SubmitOffer(&bitfinex.FundingOfferRequest{
		Type:   fundingType,
		Symbol: symbol,
		Amount: amount,
		Rate:   rate,
		Period: day,
		Hidden: false,
	})
	if err != nil {
		logger.LOG.Errorf("Funding Offer Failed : %v", err)
		return err
	}
	newOffer := fo.NotifyInfo.(*bitfinex.FundingOfferNew)
	utils.PrintWithStruct(newOffer)
	return nil
}

func (api *APIClient) CancelFundingOffer(userId, offerId int64) {
	if api.CheckRateCount() != nil {
		return
	}

	if client := api.GetClientByUserId(userId); client != nil {
		_, err := client.Funding.CancelOffer(&bitfinex.FundingOfferCancelRequest{
			Id: offerId,
		})

		if err != nil {
			logger.LOG.Errorf("Cancel offer error : %v", offerId)
		}
	}
}
