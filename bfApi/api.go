package bfApi

import (
	"log"
	"os"
	"time"

	"github.com/bitfinexcom/bitfinex-api-go/v2"
	"robot/logger"
	"robot/utils"

	"github.com/bitfinexcom/bitfinex-api-go/v2/rest"
)

var client *rest.Client

func ApiInit() {
	key := os.Getenv("API_KEY")
	secret := os.Getenv("API_SEC")
	url := os.Getenv("BFX_API_URI")
	client = rest.NewClientWithURL(url).Credentials(key, secret)
}

func TickerAction() {
	symbols := []string{bitfinex.FundingPrefix+"USD"}
	tickers, err := client.Tickers.GetMulti(symbols)

	if err != nil {
		log.Printf("getting ticker: %s", err)
	}

	utils.PrintWithStruct(tickers)
}

// 每日funding offer 利息獲得及總資產
func GetLedgers(end int64) []*bitfinex.Ledger{
	result, err := client.Ledgers.Ledgers("USD", 0, end, 500)
	if err != nil {
		logger.LOG.Errorf("getting Ledgers: %s", err)
		return nil
	}

	return result.Snapshot
}

func GetBook(precision bitfinex.BookPrecision) (bid []*bitfinex.BookUpdate, offer []*bitfinex.BookUpdate ,err error){
	book, err := client.Book.All(bitfinex.FundingPrefix+"USD", precision, 100)

	if err != nil {
		logger.LOG.Errorf("Get book list: %s", err)
		return
	}

	return book.Snapshot[0:100], book.Snapshot[100:], nil
}



func GetMatched(limit int) ([]*bitfinex.Trade, error){
	fiveMin, _ := time.ParseDuration("-2h")

	now:= time.Now()
	start:= bitfinex.Mts(now.Add(fiveMin).UnixNano()/ int64(time.Millisecond))
	end := bitfinex.Mts(now.UnixNano()/ int64(time.Millisecond))

	matchedList, err := client.Trades.PublicHistoryWithQuery(bitfinex.FundingPrefix+"USD", start,end, bitfinex.QueryLimit(limit), bitfinex.NewestFirst)

	if err != nil {
		logger.LOG.Errorf("Get Matched list: %v", err)
		return nil, err
	}

	return matchedList.Snapshot, nil
}

func GetOnOfferList () []*bitfinex.Offer{
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

func SubmitFundingOffer(symbol string, ffr bool, amount float64,rate float64, day int64) error{
	fundingType := "LIMIT"
	if ffr {
		fundingType = "FRRDELTAVAR"
	}

	fo, err := client.Funding.SubmitOffer(&bitfinex.FundingOfferRequest{
		Type: fundingType,
		Symbol: symbol,
		Amount: amount,
		Rate: rate,
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

func CancelFundingOffer(offerId int64) {
	_, err := client.Funding.CancelOffer(&bitfinex.FundingOfferCancelRequest{
		Id: offerId,
	})

	if err != nil {
		logger.LOG.Errorf("Cancel offer error : %v", offerId)
	}
}
