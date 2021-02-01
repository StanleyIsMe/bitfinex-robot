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
				if walletStatus.Type == "funding" && walletStatus.Currency == "USD" {
					updateWalletChan <- walletStatus
				}
				//if walletStatus.BalanceAvailable >= 50 && walletStatus.Type == "funding" {
				//	//wallet.Update(walletStatus.Balance, walletStatus.BalanceAvailable)
				//	updateWalletChan <- 1
				//}
			case *wallet.Snapshot:
				walletSnapshot := obj.(*wallet.Snapshot)
				for _, wallets := range walletSnapshot.Snapshot {
					if wallets.Type == "funding" && wallets.Currency == "USD"{
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

func (st *Socket) CalWalletUpdate() {
	msg := []interface{}{0, "calc", nil, [][]string{{"wallet_funding_USD"}}}
	err := st.Client.Send(context.Background(), msg)

	if err != nil {
		log.Fatal(err)
	}
}
