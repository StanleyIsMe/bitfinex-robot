package crontab

import (
	"time"

	"github.com/astaxie/beego/toolbox"
	bfApi "robot/btApi"
	"robot/lineBot"
	"robot/utils"
)

func Start() {
	// 每天09:30:30 Am
	task1 := toolbox.NewTask("放貸收穫", "20 30 9 * * *", func() error {
		// work
		earnInfoList := map[string]float64{}
		list := bfApi.GetLedgers()
		for _, data := range list {
			if data.Description == "Margin Funding Payment on wallet funding" {
				dateTime := time.Unix(data.MTS/1000, 0).Format("2006-01-02 15:04:05")
				earnInfoList[dateTime] = data.Amount
				earnInfoList["Balance"] = data.Balance
				break
			}
		}

		content, _ := utils.JsonString(earnInfoList)
		lineBot.LineSendMessage(content)
		return nil
	})

	task2 := toolbox.NewTask("機器人檢查", "0 */30 * * * *", func() error {
		// work
		lineBot.LineSendMessage("我還在工作唷")
		return nil
	})
	toolbox.AddTask("放貸收穫", task1)
	toolbox.AddTask("機器人檢查", task2)

	toolbox.StartTask()
}