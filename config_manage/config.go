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
	Day          int     `json:"天數"`
	CrazyRate    float64 `json:"瘋狂利率"`
	IncreaseRate float64 `json:"遞增利率"`
	TelegramId   int64
	SubmitOffer  bool           `json:"自動放貸"`
	InValidRate  float64        `json:"無效利率"`
	Weights      map[string]int `json:"利率計算權重"`
	//Policy func()float64 `json:"-"`
}

var Config *ConfigManage

func NewConfig() *ConfigManage {
	var once sync.Once

	if Config == nil {
		once.Do(func() {

			bottomRate, _ := strconv.ParseFloat(os.Getenv("FUNDING_BOTTOM_RATE"), 64)
			crazyRate, _ := strconv.ParseFloat(os.Getenv("FUNDING_CRAZY_RATE"), 64)
			fixedAmount, _ := strconv.ParseFloat(os.Getenv("FUNDING_FIXED_AMOUNT"), 64)
			increaseRate, _ := strconv.ParseFloat(os.Getenv("FUNDING_INCREASE_RATE"), 64)
			telegramId, _ := strconv.ParseInt(os.Getenv("TELEGRAM_MANAGE_ID"), 10, 64)
			invalidRate, _ := strconv.ParseFloat(os.Getenv("INVALID_RATE"), 64)
			submitOffer := os.Getenv("AUTO_SUBMIT_FUNDING") == "Y"

			Config = &ConfigManage{
				RWMutex:      sync.RWMutex{},
				BottomRate:   bottomRate,
				FixedAmount:  fixedAmount,
				Day:          2,
				CrazyRate:    crazyRate,
				IncreaseRate: increaseRate,
				TelegramId:   telegramId,
				SubmitOffer:  submitOffer,
				InValidRate:  invalidRate,
				Weights: map[string]int{
					"book01":   1,
					"book02":   1,
					"book03":   8,
					"book04":   10,
					"avg100":   1,
					"avg10000": 3,
				},
			}
		})
	}

	return Config
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

func (config *ConfigManage) GetInValidRate() float64 {
	config.Lock()
	defer config.Unlock()
	return config.InValidRate
}

func (config *ConfigManage) GetSubmitOffer() bool {
	config.Lock()
	defer config.Unlock()
	return config.SubmitOffer
}

func (config *ConfigManage) GetWeights() map[string]int {
	config.Lock()
	defer config.Unlock()
	return config.Weights
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

func (config *ConfigManage) SetInValidRate(rate float64) {
	config.Lock()
	defer config.Unlock()
	config.InValidRate = rate
}

func (config *ConfigManage) SetWeights(key string, increment int) {
	config.Lock()
	defer config.Unlock()
	val, ok := config.Weights[key]
	if ok && (val+increment) > 0 {
		config.Weights[key] += increment
	}
}

// 權重初始化
func (config *ConfigManage) WeightsInit() {
	config.Lock()
	defer config.Unlock()
	config.Weights = map[string]int{
		"book01":   1,
		"book02":   1,
		"book03":   8,
		"book04":   10,
		"avg100":   1,
		"avg10000": 3,
	}
}
