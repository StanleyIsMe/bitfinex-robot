package bfApi

import (
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
	fmt.Println("funding offer history", "===================================")
	snapHist, err := client.Funding.OfferHistory("fUSD")
	if err != nil {
		panic(err)
	}

	utils.PrintWithStruct(snapHist)

	fmt.Println("active credits", "===================================")
	snapCredits, err := client.Funding.Credits("fUSD")
	if err != nil {
		panic(err)
	}
	utils.PrintWithStruct(snapCredits)

	// credits history
	fmt.Println("credits history", "===================================")
	napCreditsHist, err := client.Funding.CreditsHistory("fUSD")
	if err != nil {
		panic(err)
	}

	utils.PrintWithStruct(napCreditsHist)

	// funding trades
	fmt.Println("funding trades", "===================================")
	napTradesHist, err := client.Funding.Trades("fUSD")
	if err != nil {
		panic(err)
	}
	utils.PrintWithStruct(napTradesHist)
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
func LedgersAction() {
	now:= time.Now()
	fmt.Println(now.UnixNano())
	result, err := client.Ledgers.Ledgers("USD", 0, 1580139221740, 500)
	if err != nil {
		log.Fatalf("getting orders: %s", err)
	}

	utils.PrintWithStruct(result)
}

func submitFundingOffer(symbol string, ffr bool, rate float64, day int64){
	fundingType := "LIMIT"
	if ffr {
		fundingType = "FRRDELTAVAR"
	}

	fo, err := client.Funding.SubmitOffer(&bitfinex.FundingOfferRequest{
		Type: fundingType,
		Symbol: symbol,
		Amount: 50,
		Rate: rate,
		Period: day,
		Hidden: false,
	})
	if err != nil {
		panic(err)
	}
	newOffer := fo.NotifyInfo.(*bitfinex.FundingOfferNew)
	utils.PrintWithStruct(newOffer)
}

