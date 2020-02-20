package telegramBot

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"robot/bfApi"
	"robot/config_manage"
	"robot/utils"
)

var bot *tgbotapi.BotAPI
var numericKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("利息"),
		//tgbotapi.NewKeyboardButton("放貸金額"),
		//tgbotapi.NewKeyboardButton("錢包"),
		tgbotapi.NewKeyboardButton("config"),
	),
	//tgbotapi.NewKeyboardButtonRow(
	//	tgbotapi.NewKeyboardButton("CrazyRate:0.0009"),
	//	tgbotapi.NewKeyboardButton("CrazyRate:0.00085"),
	//	tgbotapi.NewKeyboardButton("CrazyRate:0.0008"),
	//	tgbotapi.NewKeyboardButton("CrazyRate:0.00075"),
	//),
)

type Rate float64

var ActionBook = map[string]string{
	"利息": "1",
}

func BotInit() {
	if bot == nil {
		var err error
		bot, err = tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
		if err != nil {
			log.Printf("telegram-bot connect error : %v", err)
		}

		bot.Debug = true

		log.Printf("Authorized on account %s", bot.Self.UserName)
	}

}

func Listen() {
	go func() {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates, err := bot.GetUpdatesChan(u)

		if err != nil {
			log.Printf("telegram-bot update channel error : %v", err)
		}

		for update := range updates {
			if update.Message == nil { // ignore non-Message updates
				continue
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Text {
			case "open":
				msg.ReplyMarkup = numericKeyboard
				break
			case "close":
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				break
			case "config":
				content, _ := utils.JsonString(config_manage.Config)
				msg.Text = content
				break
			case "利息":
				msg.Text = GetInterestInfo()
			default:
				key, val := parseText(update.Message.Text)
				msg.Text = ReplyAction(key, val)
			}

			if msg.Text == "" {
				continue
			}

			if _, err := bot.Send(msg); err != nil {
				log.Printf("Response Telegram Reply Error : %v", err)
			}
		}
	}()
}

func SendMessage(chatId int64, text string) {
	if bot == nil {
		BotInit()
	}

	msg := tgbotapi.NewMessage(chatId, text)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Send Message Error : %v", err)
	}
}

func ServerMessage(text string) {
	if bot == nil {
		BotInit()
	}

	chatId, _ := strconv.ParseInt(os.Getenv("TELEGRAM_MANAGE_ID"), 10, 64)
	msg := tgbotapi.NewMessage(chatId, text)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Send Message Error : %v", err)
	}
}

func Close() {
	bot.StopReceivingUpdates()
}

func parseText(input string) (string, string) {
	input = strings.Replace(input, " ", "", -1)
	split := strings.Split(input, ":")

	switch len(split) {
	case 0:
		return "", ""
	case 1:
		return split[0], ""
	case 2:
		return split[0], split[1]
	default:
		return "", ""
	}
}

func ReplyAction(key, val string) (reply string) {
	if val == "" || key == "" {
		return "找不到對應動作"
	}

	reply = "執行完畢"
	switch key {
	case "CrazyRate":
		rate, _ := strconv.ParseFloat(val, 64)
		config_manage.Config.SetCrazyRate(rate)
		break
	case "IncreaseRate":
		rate, _ := strconv.ParseFloat(val, 64)
		config_manage.Config.SetIncreaseRate(rate)
		break
	case "BottomRate":
		rate, _ := strconv.ParseFloat(val, 64)
		config_manage.Config.SetBottomRate(rate)
		break
	case "FixedAmount":
		rate, _ := strconv.ParseFloat(val, 64)
		config_manage.Config.SetFixedAmount(rate)
		break
	case "InValidRate":
		rate, _ := strconv.ParseFloat(val, 64)
		config_manage.Config.SetInValidRate(rate)
		break
	case "Day":
		day, _ := strconv.Atoi(val)
		config_manage.Config.SetDay(day)
		break
	case "SubmitOffer":
		config_manage.Config.SetSubmitOffer(val == "Y" || val == "y")
		break
	default:
		reply = "找不到對應動作"
	}

	return reply
}

type DailyInterestReport struct {
	Balance float64 `json:"錢包總額"`
	TotalInterest float64 `json:"利息總額"`
	InterestList []map[string]interface{} `json:"利息清單"`
}

// 取得近十天的利息
func GetInterestInfo() string {
	report := &DailyInterestReport{}

	end := time.Now().UnixNano()/ int64(time.Millisecond)
	list := bfApi.GetLedgers(end)
	count := 0
	for len(list) > 0 {
		for _, data := range list {

			if data.Description == "Margin Funding Payment on wallet funding" {
				count++

				// 第一筆為總金額
				if count == 1 {
					report.Balance = data.Balance
				}

				report.TotalInterest += data.Amount

				if count > 10 {
					continue
				}
				earnInfo := map[string]interface{}{}
				dateTime := time.Unix(data.MTS/1000, 0).Format("2006-01-02 15:04:05")
				earnInfo["Date"] = dateTime
				earnInfo["Interest"] = data.Amount
				report.InterestList = append(report.InterestList, earnInfo)
			}
			end = data.MTS
		}

		list = bfApi.GetLedgers(end-1)
	}



	content, _ := utils.JsonString(report)
	//ServerMessage(content)
	log.Print("Get Interest Info Done")
	return content
}

