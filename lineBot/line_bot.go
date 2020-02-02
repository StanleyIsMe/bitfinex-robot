package lineBot

import (
	"log"
	//"net/http"
	"os"
	"fmt"


	"github.com/line/line-bot-sdk-go/linebot"
)

var bot *linebot.Client
func LineInit() {
	var err error
	bot, err = linebot.New(os.Getenv("Line_CHANNEL_SECRET"), os.Getenv("LINE_CHANNEL_TOKEN"))
	if err != nil {
		fmt.Printf("New Line-Bot Some thing error : %v", err)
		//SendEmail(fmt.Sprintf("New Line-Bot Some thing error : %v", err), "Line Bot Error")

	}

	if _, err := bot.ReplyMessage("", linebot.NewTextMessage("hello")).Do(); err != nil {

	}

	LineSendMessage("robot 啟動")
}

func LineSendMessage(message string) {
	// append some message to messages
	messages := []linebot.SendingMessage{linebot.NewTextMessage(message)}
	_, err := bot.PushMessage(os.Getenv("LINE_USER_ID"), messages...).Do()
	if err != nil {
		log.Printf("line send message error: %v", err)
	}
}
