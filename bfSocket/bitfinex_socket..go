package bfSocket

import (
	"context"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/fundingoffer"
	"log"
	"os"
	"robot/utils"
	"time"

	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/wallet"
	//"github.com/bitfinexcom/bitfinex-api-go/v2/rest"
	"github.com/bitfinexcom/bitfinex-api-go/v2/websocket"
)

//var socket *websocket.Client

type Socket struct {
	Client *websocket.Client
}

func NewSocket(key, secret string) *Socket {

	p := websocket.NewDefaultParameters()

	// Enable orderbook checksum verification
	p.ManageOrderbook = true
	p.ReconnectAttempts = 999999999
	p.ReconnectInterval = time.Second * 30

	p.URL = os.Getenv("BFX_WS_URI")
	//socket = websocket.NewWithParams(p).Credentials(key, secret)

	socket := &Socket{
		Client: websocket.NewWithParams(p).Credentials(key, secret),
	}

	err := socket.Client.Connect()
	if err != nil {
		log.Fatal("Key [%s] Error connecting to bitfinex web socket : ", key, err)
	}

	return socket
}

func (st *Socket) IsConnected() bool {
	return st.Client.IsConnected()
}

func (st *Socket) Close() {
	st.Client.Close()
}

func (st *Socket) Listen(updateWalletChan chan *wallet.Update) {
	go func() {
		//wallet := policy.NewWallet()
		for obj := range st.Client.Listen() {
			switch obj.(type) {
			case error:
				log.Printf("Socket error: %v", obj.(error))
			//case *bitfinex.WalletUpdate:
			case *wallet.Update:
				walletStatus := obj.(*wallet.Update)
				if walletStatus.Type == "funding" {
					updateWalletChan <- walletStatus
				}
				//if walletStatus.BalanceAvailable >= 50 && walletStatus.Type == "funding" {
				//	//wallet.Update(walletStatus.Balance, walletStatus.BalanceAvailable)
				//	updateWalletChan <- 1
				//}
			case *wallet.Snapshot:
				walletSnapshot := obj.(*wallet.Snapshot)
				for _, wallets := range walletSnapshot.Snapshot {
					if wallets.Type == "funding" {
						utils.PrintWithStruct(wallets)
						newWalletUpdate := &wallet.Update{
							Balance:          wallets.Balance,
							BalanceAvailable: wallets.BalanceAvailable,
						}
						updateWalletChan <- newWalletUpdate
					}
				}
			//case *fundingoffer.Snapshot:
			//	fundingOffer := obj.(*fundingoffer.Snapshot)
			//case *bitfinex.FundingOfferNew:
			//case *bitfinex.FundingOfferUpdate:
			//// 個人funding 交易 即時狀況
			//case *bitfinex.FundingTrade:
			//fundingTrade := obj.(*bitfinex.FundingTrade)
			//content, _ := utils.JsonString(fundingTrade)
			//lineBot.LineSendMessage(content)

			default:
				utils.PrintWithStruct(obj)
			}
		}
	}()
}

func (st *Socket) SubmitFundingOffer(symbol string, amount float64, rate float64, day int64) error {
	log.Printf("Submitting new funding offer")
	err := st.Client.SubmitFundingOffer(context.Background(), &fundingoffer.SubmitRequest{
		Type:   "LIMIT",
		Symbol: symbol,
		Amount: amount,
		Rate:   rate,
		Period: day,
		Hidden: false,
	})
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func (st *Socket) CancelFundingOffer(offerId int64) error {
	log.Printf("Submitting cancel funding offer")
	err := st.Client.SubmitFundingCancel(context.Background(), &fundingoffer.CancelRequest{
		ID: offerId,
	})
	if err != nil {
		log.Fatal(err)
	}
	return err
}

//
//func SocketInit() {
//
//	p := websocket.NewDefaultParameters()
//
//	// Enable orderbook checksum verification
//	p.ManageOrderbook = true
//	p.ReconnectAttempts = 999999999
//	p.ReconnectInterval = time.Second * 30
//
//	key := os.Getenv("API_KEY")
//	secret := os.Getenv("API_SEC")
//	//url := os.Getenv("BFX_API_URI")
//	p.URL = os.Getenv("BFX_WS_URI")
//	socket = websocket.NewWithParams(p).Credentials(key, secret)
//
//	err := socket.Connect()
//	if err != nil {
//		log.Fatal("Error connecting to bitfinex web socket : ", err)
//	}
//}
//
//func Listen(notifyChannel chan int) {
//	//ctx, cxl2 := context.WithTimeout(context.Background(), time.Second*5)
//	//defer cxl2()
//	//_, err := socket.SubscribeTicker(ctx, "fUSD")
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//
//	//ctx, cxl2 := context.WithTimeout(context.Background(), time.Second*5)
//	//defer cxl2()
//	//_, err := socket.SubscribeBook(ctx, "fUSD", bitfinex.Precision2, bitfinex.FrequencyTwoPerSecond, 25)
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//	//
//	//ctx, cxl3 := context.WithTimeout(context.Background(), time.Second*5)
//	//defer cxl3()
//	//_, err = socket.SubscribeTrades(ctx, bitfinex.FundingPrefix+"USD")
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//
//	go func() {
//		wallet := policy.NewWallet()
//		for obj := range socket.Listen() {
//			switch obj.(type) {
//			case error:
//				log.Printf("Socket error: %v", obj.(error))
//				//lineBot.LineSendMessage("Socket error")
//				//utils.SendEmail(fmt.Sprintf("channel closed: %s", obj), "robot socket error")
//			case *bitfinex.WalletUpdate:
//				walletStatus := obj.(*bitfinex.WalletUpdate)
//				if walletStatus.BalanceAvailable >= 50 && walletStatus.Type == "funding" {
//					wallet.Update(walletStatus.Balance, walletStatus.BalanceAvailable)
//					//content, _ := utils.JsonString(walletStatus)
//					//lineBot.LineSendMessage(content)
//					notifyChannel <- 1
//					//SendEmail(content, "wallet status")
//				}
//
//			case *bitfinex.FundingOfferNew:
//				//fundingStatus := obj.(*bitfinex.FundingOfferNew)
//				//content, _ := utils.JsonString(fundingStatus)
//				//lineBot.LineSendMessage(content)
//				//SendEmail(content, fmt.Sprintf("New Funding Offer :$%f ,rate: %f", fundingStatus.Amount, fundingStatus.Rate) )
//			case *bitfinex.FundingOfferUpdate:
//				//fundingStatus := obj.(*bitfinex.FundingOfferUpdate)
//				//if fundingStatus.Status == bitfinex.OfferStatusExecuted {
//				//	content, _ := utils.JsonString(fundingStatus)
//				//	lineBot.LineSendMessage(content)
//				//	//SendEmail(content, fmt.Sprintf("New Funding Executed :$%f ,rate: %f", fundingStatus.Amount, fundingStatus.Rate) )
//				//}
//				// 即時最新funding offer/bid 價況，及matched 價格
//			//case *bitfinex.Ticker:
//			//	ticker := obj.(*bitfinex.Ticker)
//			//	content, _ := utils.JsonString(ticker)
//			//	lineBot.LineSendMessage(content)
//			//case *bitfinex.Trade:
//			//	utils.PrintWithStruct(obj)
//			//	//matchedRealTime := obj.(*bitfinex.Trade)
//			//	//content, _ := utils.JsonString(ticker)
//			//	//lineBot.LineSendMessage(content)
//			// 個人funding 交易 即時狀況
//			case *bitfinex.FundingTrade:
//				//fundingTrade := obj.(*bitfinex.FundingTrade)
//				//content, _ := utils.JsonString(fundingTrade)
//				//lineBot.LineSendMessage(content)
//
//			default:
//				//utils.PrintWithStruct(obj)
//			}
//
//			//fmt.Println("MSG RECV:===============")
//			////JsonPrint(obj)
//			////fmt.Println("SPEW ==============")
//			//spew.Dump(obj)
//			////log.Printf("MSG RECV: %#v", obj)
//			//
//			//// Load the latest orderbook
//
//			//ob, _ := socket.GetOrderbook("fUSD")
//			//if ob != nil {
//			//	//utils.PrintWithStruct(ob)
//			//	//fmt.Println("Ask================")
//			//	//JsonPrint(ob.Asks())
//			//	//fmt.Println("Bids================")
//			//	utils.PrintWithStruct(ob.Bids())
//			//	//log.Printf("Orderbook asks: %v", ob.Asks())
//			//	//log.Printf("Orderbook bids: %v", ob.Bids())
//			//}
//
//			//ticker,_ := socket.GetOrderbook()
//		}
//	}()
//}
//
//func IsConnected() bool {
//	return socket.IsConnected()
//}
//
//func Close() {
//	socket.Close()
//}
