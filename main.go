package main

import (
	//"context"

	//"time"

	"log"
	//"os/signal"

	//"net/http"
	"os"
	"os/signal"



	//"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
	//"github.com/davecgh/go-spew/spew"
	//"github.com/bitfinexcom/bitfinex-api-go/v2"
	"robot/lineBot"
	"robot/bfSocket"
	"robot/btApi"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	bfApi.ApiInit()
	//bfApi.LedgersAction()
	//return
	bfSocket.SocketInit()
	lineBot.LineInit()
	// subscribe to BTCUSD book
	//ctx, cxl2 := context.WithTimeout(context.Background(), time.Second*5)
	//defer cxl2()
	//_, err = c.SubscribeTicker(ctx, "fUSD")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//// subscribe to BTCUSD trades
	//ctx, cxl3 := context.WithTimeout(context.Background(), time.Second*5)
	//defer cxl3()
	//_, err = c.SubscribeTrades(ctx, "fUSD")
	//if err != nil {
	//	log.Fatal(err)
	//}


	bfSocket.Listen()


	done := make(chan bool, 1)
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill)
	go func() {
		<-interrupt
		bfSocket.Close()
		lineBot.LineSendMessage("robot 結束")
		done <- true
		os.Exit(0)
	}()
	<-done
}





func Policy() {
	// 2天
	// 5天
	// 11天
	// 30天
}