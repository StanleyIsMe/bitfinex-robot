package config_manage

import (
	"os"
	"strconv"
	"sync"
)

type ConfigManage struct {
	sync.RWMutex

	BottomRate   float64 `json:"最低利率"`
	FixedAmount  float64 `json:"固定放貸額"`
	Day          int `json:"天數"`
	CrazyRate    float64 `json:"瘋狂利率"`
	IncreaseRate float64 `json:"遞增利率"`
	TelegramId   int64
	SubmitOffer  bool `json:"自動放貸"`
	Policy func()float64
}

var config *ConfigManage
func NewConfig() *ConfigManage {
	var once sync.Once

	if config == nil {
		once.Do(func() {

			bottomRate, _ := strconv.ParseFloat(os.Getenv("FUNDING_BOTTOM_RATE"), 64)
			crazyRate, _ := strconv.ParseFloat(os.Getenv("FUNDING_CRAZY_RATE"), 64)
			fixedAmount, _ := strconv.ParseFloat(os.Getenv("FUNDING_FIXED_AMOUNT"), 64)
			increaseRate, _ := strconv.ParseFloat(os.Getenv("FUNDING_INCREASE_RATE"), 64)
			telegramId, _ := strconv.ParseInt(os.Getenv("TELEGRAM_MANAGE_ID"), 10, 64)
			submitOffer := os.Getenv("AUTO_SUBMIT_FUNDING") == "Y"

			config = &ConfigManage{
				RWMutex:      sync.RWMutex{},
				BottomRate:   bottomRate,
				FixedAmount:  fixedAmount,
				Day:          2,
				CrazyRate:    crazyRate,
				IncreaseRate: increaseRate,
				TelegramId:   telegramId,
				SubmitOffer:  submitOffer,
			}
		})
	}


	return config
}

func (config *ConfigManage) GetBottomRate() float64 {
	config.Lock()
	defer config.Unlock()
	return config.BottomRate
}

func (config *ConfigManage) GetFixedAmount() float64 {
	config.Lock()
	defer config.Unlock()
	return config.FixedAmount
}

func (config *ConfigManage) GetDay() int {
	config.Lock()
	defer config.Unlock()
	return config.Day
}

func (config *ConfigManage) GetCrazyRate() float64 {
	config.Lock()
	defer config.Unlock()
	return config.CrazyRate
}

func (config *ConfigManage) GetIncreaseRate() float64 {
	config.Lock()
	defer config.Unlock()
	return config.IncreaseRate
}

func (config *ConfigManage) GetSubmitOffer() bool {
	config.Lock()
	defer config.Unlock()
	return config.SubmitOffer
}

func (config *ConfigManage) SetBottomRate(rate float64) {
	config.Lock()
	defer config.Unlock()
	config.BottomRate = rate
}

func (config *ConfigManage) SetFixedAmount(amount float64) {
	config.Lock()
	defer config.Unlock()
	config.FixedAmount = amount
}

func (config *ConfigManage) SetDay(day int) {
	config.Lock()
	defer config.Unlock()
	config.Day = day
}

func (config *ConfigManage) SetCrazyRate(rate float64) {
	config.Lock()
	defer config.Unlock()
	config.CrazyRate = rate
}

func (config *ConfigManage) SetIncreaseRate(rate float64) {
	config.Lock()
	defer config.Unlock()
	config.IncreaseRate = rate
}

func (config *ConfigManage) SetSubmitOffer(submit bool) {
	config.Lock()
	defer config.Unlock()
	config.SubmitOffer = submit
}
