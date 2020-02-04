package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bitfinexcom/bitfinex-api-go/v2"
	"github.com/joho/godotenv"
	"robot/bfSocket"
	"robot/btApi"
	"robot/crontab"
	"robot/lineBot"
	"robot/policy"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	lineBot.LineInit()
	bfApi.ApiInit()
	policy.PolicyInit()
	//rate := policy.TrackBookPrice()
	//rate2 := policy.TrackMatchPrice()
	//fmt.Println("=================================", rate, rate2)
	//
	//os.Exit(1)
	bfSocket.SocketInit()
	crontab.Start()

	notifyChannel := make(chan int)

	go submitFunding(notifyChannel)
	bfSocket.Listen(notifyChannel)

	//os.Exit(0)
	done := make(chan bool, 1)
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill)
	go func() {
		<-interrupt
		lineBot.LineSendMessage("robot 結束")
		bfSocket.Close()
		done <- true
		os.Exit(0)
	}()
	<-done
}

func submitFunding(notifyChannel <-chan int) {
	wallet := policy.NewWallet()
	for j := range notifyChannel {

		// 放貸天數
		day := 2
		// 計算放貸利率
		rate := policy.TrackMatchPrice()

		if rate <= 0.0002 {
			log.Println("計算結果低於: ", rate)
			return
		}
		log.Printf("Calculate Rate : %v, sign %v", rate, j)

		if os.Getenv("AUTO_SUBMIT_FUNDING") == "Y" {
			for wallet.BalanceAvailable >= 50 {
				if rate >= policy.MyRateController.CrazyRate {
					day = 30
				}

				amount := wallet.GetAmount(50)
				err := bfApi.SubmitFundingOffer(bitfinex.FundingPrefix+"USD", false, amount, rate, int64(day))
				if err != nil {
					lineBot.LineSendMessage(err.Error())
					break
				}
				rate += policy.MyRateController.IncreaseRate
			}
		}
	}
}
