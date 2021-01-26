package bfApi

import (
	"context"
	"errors"
	"fmt"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/book"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/common"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/fundingoffer"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/ledger"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/trade"
	"log"
	"os"
	"sync"
	"time"

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
	APIOnce.Do(func() {
		url := os.Getenv("BFX_API_URI")
		pubClient := rest.NewClientWithURL(url).Credentials(os.Getenv("API_KEY"), os.Getenv("API_SEC"))
		APIClientInstance = &APIClient{
			rateCount:  10,
			ClientList: make(map[int64]*rest.Client, 0),
			PublicClient:pubClient,
		}

		APIClientInstance.ctx, APIClientInstance.cancel = context.WithCancel(context.Background())
		go APIClientInstance.LoopCalculateRateLimit()
	})
	return APIClientInstance
}

func (api *APIClient) LoopCalculateRateLimit() {
loop:
	for {
		select {
		case <-time.After(1 * time.Minute):
			api.Lock()
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
	log.Printf("API Rate Count [%d]", api.rateCount)

	if api.rateCount >= RateLimit {
		log.Println("Reached API Rate limit")
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
		if _, err := tempClient.Wallet.Wallet(); err != nil {
			logger.LOG.Errorf("UserId [%d] Bitfinex Api Fail %v", userId, err)
			return false
		}

		log.Printf("UserId [%d] Bitfinex Api Success", userId)
		api.ClientList[userId] = tempClient
		return true
	}
	return true
}

// 每日funding offer 利息獲得及總資產
func (api *APIClient) GetLedgers(userId, end int64) []*ledger.Ledger {
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

func (api *APIClient) GetBook(precision common.BookPrecision) (bid []*book.Book, offer []*book.Book, err error) {
	if api.CheckRateCount() != nil {
		return
	}

	book, err := api.PublicClient.Book.All(common.FundingPrefix+"USD", precision, 100)

	if err != nil {
		logger.LOG.Errorf("Get book list: %s", err)
		return
	}

	return book.Snapshot[0:100], book.Snapshot[100:], nil
}

func (api *APIClient) GetMatched(limit int) ([]*trade.Trade, error) {
	if api.CheckRateCount() != nil {
		return nil, nil
	}

	fiveMin, _ := time.ParseDuration("-2h")

	now := time.Now()
	start := common.Mts(now.Add(fiveMin).UnixNano() / int64(time.Millisecond))
	end := common.Mts(now.UnixNano() / int64(time.Millisecond))

	matchedList, err := api.PublicClient.Trades.PublicHistoryWithQuery(common.FundingPrefix+"USD", start, end, common.QueryLimit(limit), common.NewestFirst)

	if err != nil {
		logger.LOG.Errorf("Get Matched list: %v", err)
		return nil, err
	}

	return matchedList.Snapshot, nil
}

func (api *APIClient) GetOnOfferList(userId int64) []*fundingoffer.Offer {
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

		log.Printf("GetOnOfferList error : %v", err)
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

	fo, err := client.Funding.SubmitOffer(&fundingoffer.SubmitRequest{
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
	newOffer := fo.NotifyInfo.(*fundingoffer.Snapshot)
	utils.PrintWithStruct(newOffer)
	return nil
}

func (api *APIClient) CancelFundingOffer(userId, offerId int64) {
	if api.CheckRateCount() != nil {
		return
	}

	if client := api.GetClientByUserId(userId); client != nil {
		_, err := client.Funding.CancelOffer(&fundingoffer.CancelRequest{
			ID: offerId,
		})

		if err != nil {
			logger.LOG.Errorf("Cancel offer error : %v", offerId)
		}
	}
}
