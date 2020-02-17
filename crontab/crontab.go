package crontab

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"


	"github.com/astaxie/beego/toolbox"
	"robot/bfSocket"
	bfApi "robot/btApi"
	"robot/telegramBot"
	"robot/utils"
)
type DailyInterestReport struct {
	Balance float64 `json:"錢包總額"`
	TotalInterest float64 `json:"利息總額"`
	InterestList []map[string]interface{} `json:"利息清單"`
}

func Start() {
	// 每天09:31:0 Am
	task1 := toolbox.NewTask("放貸收穫", "0 35 9 * * *", func() error {
		// work
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

					earnInfo := map[string]interface{}{}
					dateTime := time.Unix(data.MTS/1000, 0).Format("2006-01-02 15:04:05")
					earnInfo["Date"] = dateTime
					earnInfo["Interest"] = data.Amount
					report.TotalInterest += data.Amount
					report.InterestList = append(report.InterestList, earnInfo)
				}

				end = data.MTS
			}
			list = bfApi.GetLedgers(end)
		}



		content, _ := utils.JsonString(report)
		utils.SendEmail(content, "利息日報")
		telegramBot.ServerMessage(content)
		log.Print("Daily Interest Report Is Done")
		//content, _ := utils.JsonString(earnInfoList)
		//lineBot.LineSendMessage(content)
		return nil
	})

	task2 := toolbox.NewTask("機器人檢查", "0 0 */4 * * *", func() error {
		// work
		//lineBot.LineSendMessage("我還在工作唷")
		//utils.SendEmail("我有在工作拉 請放心", "Robot on working")
		telegramBot.ServerMessage("我有在工作拉 請放心")
		return nil
	})

	task3 := toolbox.NewTask("Bitfinex socket validator", "0 0 * * * *", func() error {
		if !bfSocket.IsConnected() {
			log.Fatalf("Bitfinex Socket Connected Failed")
		}
		log.Println("Bitfinex Socket On Connected")
		return nil
	})

	task4 := toolbox.NewTask("Prevent Heroku Sleep", "0 */10 * * * *", func() error {
		log.Println("Heroku Wake Up")
		res, err := http.Get("https://bf-robot.herokuapp.com/")
		if err != nil {
			log.Println(err)
		}
		defer res.Body.Close()
		sitemap, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println(err)
		}
		log.Printf("%s", sitemap)
		return nil
	})

	toolbox.AddTask("放貸收穫", task1)
	toolbox.AddTask("機器人檢查", task2)
	toolbox.AddTask("Bitfinex socket validator", task3)
	toolbox.AddTask("Prevent Heroku Sleep", task4)

	toolbox.StartTask()
}