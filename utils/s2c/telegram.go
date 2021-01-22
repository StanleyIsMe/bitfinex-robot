package s2c

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
	"sync"
)

var telegramInstance *tgbotapi.BotAPI
var telegramOnce sync.Once

func NewTgMessage() {
	telegramOnce.Do(func() {
		var err error
		telegramInstance, err = tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
		if err != nil {
			log.Printf("telegram-bot connect error : %v", err)
		}

		telegramInstance.Debug = true
		log.Printf("Authorized on account %s", telegramInstance.Self.UserName)
	})

}

func SendMessage(chatId int64, text string) {
	if telegramInstance == nil {
		NewTgMessage()
	}

	msg := tgbotapi.NewMessage(chatId, text)
	if _, err := telegramInstance.Send(msg); err != nil {
		log.Printf("Send Message Error : %v", err)
	}
}