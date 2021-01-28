package handler

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"robot/logger"
	"robot/user"
	"robot/utils"
	"strconv"
)

func RegisterHandle(telegramId int64, token, secret string) (response string) {
	defer func() {
		if err := recover(); err != nil {
			logger.LOG.Errorf("RegisterHandle Panic : %v", err)
			response = "註冊失敗"
		}
	}()

	userPool := user.GetInstance()
	err := userPool.RegisterUser(telegramId, token, secret)
	if err != nil {
		errMessage := fmt.Sprintf("TelegramId [%d], Token [%s], Sec [%s] 註冊失敗: %v", telegramId, token, secret, err)
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
				"key": key,
				"value": value,
			}).Errorf("UpdateConfigHandle Panic : %v", err)
			reply = "操作失敗"
		}
	}()

	user := user.GetInstance().GetUserById(telegramId)
	if user == nil {
		return "用戶未註冊"
	}
	config := user.Config
	reply = "成功"
	switch key {
	case "IncreaseRate":
		rate, err := strconv.ParseFloat(value, 64)
		if err != nil {
			panic(err)
		}
		config.SetIncreaseRate(rate)
		break
	case "BottomRate":
		rate, err := strconv.ParseFloat(value, 64)
		if err != nil {
			panic(err)
		}
		config.SetBottomRate(rate)
		break
	case "FixedAmount":
		rate, err := strconv.ParseFloat(value, 64)
		if err != nil {
			panic(err)
		}
		config.SetFixedAmount(rate)
		break
	case "InValidRate":
		rate, err := strconv.ParseFloat(value, 64)
		if err != nil {
			panic(err)
		}

		config.SetInValidRate(rate)
		break
	case "Day":
		day, err := strconv.Atoi(value)
		if err != nil {
			panic(err)
		}

		config.SetDay(day)
		break
	case "SubmitOffer":
		config.SetSubmitOffer(value == "Y" || value == "y")
		break
	case "CrazyDayRange":
		config.SetCrazyDayRange(value)
	case "AutoCancelTime":
		cancelTime, err :=strconv.ParseInt(value,10, 64)
		if err != nil {
			panic(err)
		}

		config.SetAutoCancelTime(cancelTime)
	default:
		reply = "找不到對應動作"
	}
	return reply
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