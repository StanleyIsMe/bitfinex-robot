package handler

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"robot/logger"
	"robot/model"
	"robot/user"
	"robot/utils"
	"strconv"
)

func RegisterHandle(request model.RegisterRequest) (response string) {
	defer func() {
		if err := recover(); err != nil {
			logger.LOG.Errorf("RegisterHandle Panic : %v", err)
			response = "註冊失敗"
		}
	}()

	userPool := user.GetInstance()
	err := userPool.RegisterUser(request)
	if err != nil {
		errMessage := fmt.Sprintf("TelegramId [%d], Token [%s], Sec [%s] 註冊失敗: %v", request.UserId, request.Token, request.Sec, err)
		logger.LOG.Errorf(errMessage)
		return errMessage
	}
	return "註冊成功"
}

func CalculateRateHandle(telegramId int64) string {
	defer func() {
		if err := recover(); err != nil {
			logger.LOG.Errorf("RegisterHandle Panic : %v", err)
		}
	}()

	user := user.GetInstance().GetUserById(telegramId)
	if user == nil {
		return "用戶未註冊"
	}
	rate := user.GetFundingRate()
	return strconv.FormatFloat(rate, 'f', 10, 64)
}

func UpdateConfigHandle(telegramId int64, key, value string) (reply string) {
	defer func() {
		if err := recover(); err != nil {
			logger.LOG.WithFields(logrus.Fields{
				"key":   key,
				"value": value,
			}).Errorf("UpdateConfigHandle Panic : %v", err)
			reply = "操作失敗"
		}
	}()

	member := user.GetInstance().GetUserById(telegramId)
	if member == nil {
		return "用戶未註冊"
	}
	config := member.Config
	reply = "成功"
	switch key {
	case "increase_rate":
		rate, err := strconv.ParseFloat(value, 64)
		if err != nil {
			panic(err)
		}
		config.SetIncreaseRate(rate)
		break
	case "bottom_rate":
		rate, err := strconv.ParseFloat(value, 64)
		if err != nil {
			panic(err)
		}
		config.SetBottomRate(rate)
		break
	case "fixed_amount":
		rate, err := strconv.ParseFloat(value, 64)
		if err != nil {
			panic(err)
		}
		config.SetFixedAmount(rate)
		break

	case "Day":
		day, err := strconv.Atoi(value)
		if err != nil {
			panic(err)
		}

		config.SetDay(day)
		break
	case "submit_offer":
		config.SetSubmitOffer(value == "Y" || value == "y")
		break
	case "crazy_day_range":
		config.SetCrazyDayRange(value)
	case "auto_cancel_time":
		cancelTime, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			panic(err)
		}

		config.SetAutoCancelTime(cancelTime)
	case "notify_rate":
		rate, err := strconv.ParseFloat(value, 64)
		if err != nil {
			panic(err)
		}
		config.SetNotifyRate(rate)
		break
	default:
		return "找不到對應動作"
	}

	if err := user.GetInstance().UpdateById(telegramId); err != nil {
		return "操作失敗"
	}
	return
}

func LookConfig(telegramId int64) string {
	defer func() {
		if err := recover(); err != nil {
			logger.LOG.Errorf("LookConfig Panic : %v", err)
		}
	}()

	if user := user.GetInstance().GetUserById(telegramId); user != nil {
		response, _ := utils.JsonString(user.Config)
		return response
	} else {
		return "用戶未註冊"
	}
}

func GetInterest(telegramId int64) string {
	defer func() {
		if err := recover(); err != nil {
			logger.LOG.Errorf("GetInterest Panic : %v", err)
		}
	}()

	user := user.GetInstance().GetUserById(telegramId)
	if user == nil {
		return "用戶未註冊"
	}

	result := user.GetInterest()
	content, _ := utils.JsonString(result)
	return content
}

func Quit(telegramId int64) string {
	defer func() {
		if err := recover(); err != nil {
			logger.LOG.Errorf("Quit Panic : %v", err)
		}
	}()

	userController := user.GetInstance()
	user := userController.GetUserById(telegramId)
	if user == nil {
		return "用戶未註冊"
	}
	if err := userController.KillUser(telegramId); err != nil {
		logger.LOG.Errorf("User %d Quit Error : %v", telegramId, err)
		return "退出失敗"
	}
	return "退出成功"
}

func Kill(telegramId int64, arg string) string {
	defer func() {
		if err := recover(); err != nil {
			logger.LOG.Errorf("Kill Panic : %v", err)
		}
	}()

	killUserId, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return "參數錯誤"
	}

	manageUserId, _ := strconv.ParseInt(os.Getenv("TELEGRAM_MANAGE_ID"), 10, 64)
	if telegramId != manageUserId {
		return "非管理者無法踢除"
	}

	if telegramId == killUserId {
		return "無法踢除自己"
	}

	userController := user.GetInstance()
	if err := userController.KillUser(killUserId); err != nil {
		logger.LOG.Errorf("User %d Kill Error : %v", killUserId, err)
		return "踢除失敗"
	}
	return "踢除成功"
}

func Start(telegramId int64) string {
	defer func() {
		if err := recover(); err != nil {
			logger.LOG.Errorf("GetInterest Panic : %v", err)
		}
	}()

	user := user.GetInstance().GetUserById(telegramId)
	if user == nil {
		return "用戶未註冊"
	}
	user.StartActive()
	return "成功"
}

func Stop(telegramId int64) string {
	defer func() {
		if err := recover(); err != nil {
			logger.LOG.Errorf("GetInterest Panic : %v", err)
		}
	}()

	user := user.GetInstance().GetUserById(telegramId)
	if user == nil {
		return "用戶未註冊"
	}
	user.StopActive()
	return "成功"
}

func Wallets(telegramId int64) string {
	defer func() {
		if err := recover(); err != nil {
			logger.LOG.Errorf("GetInterest Panic : %v", err)
		}
	}()

	user := user.GetInstance().GetUserById(telegramId)
	if user == nil {
		return "用戶未註冊"
	}
	utils.PrintWithStruct(user.API.Wallets(user.TelegramId))
	return "成功"
}
