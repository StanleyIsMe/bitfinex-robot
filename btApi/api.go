package bfApi

import (
	"context"
	"encoding/json"
	//"github.com/davecgh/go-spew/spew"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/bitfinexcom/bitfinex-api-go/v2"
	"robot/lineBot"
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

func FundingAction() {
	// 還未matched的funding offer
	fmt.Println("active funding offers", "===================================")
	snap, err := client.Funding.Offers("fUSD")
	if err != nil {
		panic(err)
	}

	utils.PrintWithStruct(snap)




	// matched funding offer history
	//fmt.Println("funding offer history", "===================================")
	//snapHist, err := client.Funding.OfferHistory("fUSD")
	//if err != nil {
	//	panic(err)
	//}
	//
	//utils.PrintWithStruct(snapHist)

	//fmt.Println("active credits", "===================================")
	//snapCredits, err := client.Funding.Credits("fUSD")
	//if err != nil {
	//	panic(err)
	//}
	//utils.PrintWithStruct(snapCredits)



	// my funding matched trades
	//fmt.Println("funding trades", "===================================")
	//napTradesHist, err := client.Funding.Trades("fUSD")
	//if err != nil {
	//	panic(err)
	//}
	//utils.PrintWithStruct(napTradesHist)
}

func PositionsAction(){
	// get active positions
	positions, err := client.Positions.All()
	if err != nil {
		log.Printf("getting wallet %s", err)
	}
	if positions != nil {
		for _, p := range positions.Snapshot {
			fmt.Println(p)
		}
	}

}

func StatsAction() {
	pLStats, err := client.Stats.PositionLast("tBTCUSD", bitfinex.Long)
	if err != nil {
		log.Printf("getting getting last position stats: %s", err)
	}
	utils.PrintWithStruct(pLStats)

	pHStats, err := client.Stats.PositionHistory("fUSD", bitfinex.Long)
	if err != nil {
		log.Printf("getting getting last position stats: %s", err)
	}
	utils.PrintWithStruct(pHStats)

	scsStats, err := client.Stats.SymbolCreditSizeLast("fUSD", "tBTCUSD")
	if err != nil {
		log.Printf("getting getting last position stats: %s", err)
	}
	utils.PrintWithStruct(scsStats)

	scsHistStats, err := client.Stats.SymbolCreditSizeHistory("fUSD", "tBTCUSD")
	if err != nil {
		log.Printf("getting getting last position stats: %s", err)
	}
	utils.PrintWithStruct(scsHistStats)

	fStats, err := client.Stats.FundingLast("fUSD")
	if err != nil {
		log.Printf("getting getting last position stats: %s", err)
	}
	utils.PrintWithStruct(fStats)

	fhStats, err := client.Stats.FundingHistory("fUSD")
	if err != nil {
		log.Printf("getting getting last position stats: %s", err)
	}
	utils.PrintWithStruct(fhStats)
}
//
func TickerAction() {
	symbols := []string{bitfinex.FundingPrefix+"USD"}
	tickers, err := client.Tickers.GetMulti(symbols)

	if err != nil {
		log.Printf("getting ticker: %s", err)
	}

	utils.PrintWithStruct(tickers)
}

// 每日funding offer 利息獲得及總資產
func GetLedgers() []*bitfinex.Ledger{
	now:= time.Now()
	end := now.UnixNano()/ int64(time.Millisecond)

	result, err := client.Ledgers.Ledgers("USD", 0, end, 500)
	if err != nil {
		log.Printf("getting Ledgers: %s", err)
		return nil
	}

	return result.Snapshot
}
type BookInfo struct {
	ID          int64       // the book update ID, optional
	Symbol      string      // book symbol
	Price       float64     // updated price
	PriceJsNum  json.Number // update price as json.Number
	Count       int64       // updated count, optional
	Amount      float64     // updated amount
	AmountJsNum json.Number // update amount as json.Number
	//Side        OrderSide   // side
	//Action      BookAction  // action (add/remove)
}

func GetBook(precision bitfinex.BookPrecision) (bid []*bitfinex.BookUpdate, offer []*bitfinex.BookUpdate ,err error){
	book, err := client.Book.All(bitfinex.FundingPrefix+"USD", precision, 100)

	if err != nil {
		log.Printf("Get book list: %s", err)
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
		log.Printf("Get Matched list: %v", err)
		return nil, err
	}

	return matchedList.Snapshot, nil
}
type LoopOnOffer struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
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
			debug.Stack()
			log.Printf("on offer error : %v", err)
			object.wg.Done()
		}
	}()

	object.wg.Add(1)
loop:
	for {
		select {
		case <-time.After(1 * time.Minute):
			now := time.Now().Add(-15 * time.Minute).Unix()
			snap, err := client.Funding.Offers("fUSD")
			if err != nil {
				log.Printf("GetOnOfferList error : %v", err)
			}

			if snap != nil {
				for _, offer := range snap.Snapshot {
					if now > (offer.MTSCreated/1000) {
						_, err := client.Funding.CancelOffer(&bitfinex.FundingOfferCancelRequest{
							Id: offer.ID,
						})

						if err != nil {
							log.Printf("Cancel offer error : %v", offer.ID)
						}
						lineBot.LineSendMessage(fmt.Sprintf("單號:%d Rate: %f Day: %d ,..超過30分鐘未撮合", offer.ID, offer.Rate, offer.Period))
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
		log.Printf("Funding Offer Failed : %v", err)
		return err
	}
	newOffer := fo.NotifyInfo.(*bitfinex.FundingOfferNew)
	utils.PrintWithStruct(newOffer)
	return nil
}

