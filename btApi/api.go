package bfApi

import (
	"encoding/json"
	//"github.com/davecgh/go-spew/spew"
	"fmt"
	"log"
	"os"
	"time"
	"robot/utils"
	"github.com/bitfinexcom/bitfinex-api-go/v2"
	//"github.com/bitfinexcom/bitfinex-api-go/v1"
	//"github.com/bitfinexcom/bitfinex-api-go/v1"
	//"github.com/bitfinexcom/bitfinex-api-go/v2/rest"
	//"github.com/bitfinexcom/bitfinex-api-go/v2"
	"github.com/bitfinexcom/bitfinex-api-go/v2/rest"
)

var client *rest.Client
func ApiInit() {


	key := os.Getenv("API_KEY")
	secret := os.Getenv("API_SEC")
	url := os.Getenv("BFX_API_URI")
	client = rest.NewClientWithURL(url).Credentials(key, secret)

	//wallets, err := client.Wallet.Wallet()
	//if err != nil {
	//	log.Fatalf("getting wallet %s", err)
	//}
	//fmt.Println(wallets)
	//spew.Dump(wallets)

	//if wallets != nil {
	//
	//	JsonPrint(wallets)
	//
	//}

	//FundingAction()

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
		log.Fatalf("getting wallet %s", err)
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
		log.Fatalf("getting getting last position stats: %s", err)
	}
	utils.PrintWithStruct(pLStats)

	pHStats, err := client.Stats.PositionHistory("fUSD", bitfinex.Long)
	if err != nil {
		log.Fatalf("getting getting last position stats: %s", err)
	}
	utils.PrintWithStruct(pHStats)

	scsStats, err := client.Stats.SymbolCreditSizeLast("fUSD", "tBTCUSD")
	if err != nil {
		log.Fatalf("getting getting last position stats: %s", err)
	}
	utils.PrintWithStruct(scsStats)

	scsHistStats, err := client.Stats.SymbolCreditSizeHistory("fUSD", "tBTCUSD")
	if err != nil {
		log.Fatalf("getting getting last position stats: %s", err)
	}
	utils.PrintWithStruct(scsHistStats)

	fStats, err := client.Stats.FundingLast("fUSD")
	if err != nil {
		log.Fatalf("getting getting last position stats: %s", err)
	}
	utils.PrintWithStruct(fStats)

	fhStats, err := client.Stats.FundingHistory("fUSD")
	if err != nil {
		log.Fatalf("getting getting last position stats: %s", err)
	}
	utils.PrintWithStruct(fhStats)
}
//
func TickerAction() {
	symbols := []string{bitfinex.FundingPrefix+"USD"}
	tickers, err := client.Tickers.GetMulti(symbols)

	if err != nil {
		log.Fatalf("getting ticker: %s", err)
	}

	utils.PrintWithStruct(tickers)
}

// 每日funding offer 利息獲得及總資產
func GetLedgers() []*bitfinex.Ledger{
	now:= time.Now()
	end := now.UnixNano()/ int64(time.Millisecond)

	result, err := client.Ledgers.Ledgers("USD", 0, end, 500)
	if err != nil {
		log.Fatalf("getting Ledgers: %s", err)
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
		log.Fatalf("getting book: %s", err)
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
		log.Fatalf("getting matched list: %v", err)
		return nil, err
	}

	return matchedList.Snapshot, nil
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

