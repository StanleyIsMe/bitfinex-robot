package main

import (
	//"context"

	//"time"


	"log"


	//"net/http"
	"os"
	"os/signal"

	//"os/signal"

	//"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
	"robot/bfSocket"
	"robot/crontab"
	"robot/policy"


	"robot/btApi"
	//"github.com/davecgh/go-spew/spew"
	//"github.com/bitfinexcom/bitfinex-api-go/v2"
	"robot/lineBot"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	lineBot.LineInit()
	bfApi.ApiInit()
	bfSocket.SocketInit()
	crontab.Start()

	policy.PolicyInit()



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
		rate, day, err := policy.Policy()
		log.Printf("Calculate Rate : %v, sign %v", rate, j)
		if err != nil {
			log.Fatal("Policy error ",err)
		}
		//
		//fmt.Println("rate:",rate, " day:", day)
		if os.Getenv("AUTO_SUBMIT_FUNDING") == "Y" {
			for wallet.BalanceAvailable >= 50 {
				amount := wallet.GetAmount(50)
				err := bfApi.SubmitFundingOffer("fUSD", false, amount, rate, int64(day))
				if err != nil {
					lineBot.LineSendMessage(err.Error())
					break
				}
				rate += policy.MyRateController.IncreaseRate
			}


		}
	}
}


