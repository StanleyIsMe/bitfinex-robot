package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bitfinexcom/bitfinex-api-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"robot/bfApi"
	"robot/bfSocket"
	"robot/config_manage"
	"robot/crontab"
	"robot/logger"
	"robot/loop"
	"robot/policy"
	"robot/telegramBot"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config_manage.NewConfig()
	logger.InitLogger()
	policy.InitPolicy()
	telegramBot.BotInit()
	telegramBot.Listen()
	bfApi.ApiInit()
	bfSocket.SocketInit()
	crontab.Start()

	// 監聽超過15分鐘未matched的單
	offerLoop := loop.NewLoopOnOffer()

	notifyChannel := make(chan int)
	go submitFunding(notifyChannel)
	bfSocket.Listen(notifyChannel)

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "deploy ok",
		})
	})

	srv := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LOG.Errorf("listen: %s\n", err)
		}
	}()

	//os.Exit(0)
	done := make(chan bool, 1)
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		<-interrupt
		telegramBot.ServerMessage("Robot Close")
		srv.Close()
		telegramBot.Close()
		offerLoop.ShutDown()
		bfSocket.Close()
		close(done)
		os.Exit(0)
	}()
	<-done
}

func submitFunding(notifyChannel <-chan int) {
	wallet := policy.NewWallet()

	for j := range notifyChannel {

		if wallet.BalanceAvailable < 50 {
			continue
		}

		// 放貸天數
		day := config_manage.Config.GetDay()
		// 計算放貸利率
		rate := policy.TrackMatchPrice2()

		if rate <= config_manage.Config.InValidRate {
			log.Println("計算結果低於: ", rate)
			return
		}
		logger.LOG.Infof("Calculate Rate : %v, sign %v", rate, j)

		if config_manage.Config.GetSubmitOffer() {
			for wallet.BalanceAvailable >= 50 {
				if rate >= config_manage.Config.GetCrazyRate() {
					day = 30
				}
				fixedAmount := config_manage.Config.GetFixedAmount()
				amount := wallet.GetAmount(fixedAmount)
				err := bfApi.SubmitFundingOffer(bitfinex.FundingPrefix+"USD", false, amount, rate, int64(day))
				if err != nil {
					logger.LOG.Errorf("Submit Offer Error: %v", err)
					break
				}
				rate += config_manage.Config.GetIncreaseRate()
			}
		}
	}
}
