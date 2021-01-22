package handler

import (
	"fmt"
	"robot/logger"
	"robot/user"
)

func RegisterHandle(telegramId int64, token, secret string) string {
	userPool := user.GetInstance()
	err := userPool.RegisterUser(telegramId, token, secret)
	if err != nil {
		errMessage := fmt.Sprintf("TelegramId [%d], Token [%s], Sec [%s] 註冊失敗: %v", telegramId, token, secret, err)
		logger.LOG.Errorf(errMessage)
		return errMessage
	}
	return "註冊成功"
}
