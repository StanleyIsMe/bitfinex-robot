package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"robot/config_manage"
	"robot/user"
	"robot/utils/s2c"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"robot/logger"
	"robot/telegramBot"
	"robot/utils/redis"
)

//func demo() {
//	pool := user.NewPool(10)
//	pool.RegisterUser(123,"","")
//	result := pool.GetAllUser()
//	os.Exit(1)
//}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	logger.InitLogger()
	config_manage.NewConfig()
	//utils.InitWorkerPool()
	redis.Init()
	//bfApi.ApiInit()
	user.NewPool(5)
	//demo()

	s2c.NewTgMessage()
	//policy.InitPolicy()
	telegramBot.BotInit()
	telegramBot.Listen()
	//bfApi.ApiInit()
	//bfSocket.SocketInit()
	//crontab.Start()


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
		//offerLoop.ShutDown()
		//bfSocket.Close()
		close(done)
		os.Exit(0)
	}()
	<-done
}
