package user

import (
	"github.com/fatih/structs"
	"os"
	"strconv"
	"strings"
	"sync"
)

type ConfigManage struct {
	sync.RWMutex `json:"-"`

	BottomRate     float64         `json:"bottom_rate"`   //最低利率"
	FixedAmount    float64         `json:"fixed_amount"`  //`json:"固定放貸額"`
	Day            int             `json:"day"`           //"天數"`
	CrazyRate      float64         `json:"crazy_rate"`    //瘋狂利率"`
	IncreaseRate   float64         `json:"increase_rate"` //遞增利率"`
	TelegramId     int64           `json:"telegram_id"`
	SubmitOffer    bool            `json:"submit_offer"` //自動放貸"`
	InvalidRate    float64         `json:"invalid_rate"` //無效利率"`
	CrazyDayRange  map[int]float64 `json:"crazy_day_range"`
	Weights        map[string]int  `json:"weights"` //利率計算權重"`
	NotifyRate     float64         `json:"notify_rate"`
	AutoCancelTime int64           `json:"auto_cancel_time"`
	OfficialMaxDay int             `json:"official_max_day"`
	OfficialMinDay int             `json:"official_min_day"`
}

//var Config *ConfigManage
// default system config
func NewConfig() *ConfigManage {
	bottomRate, _ := strconv.ParseFloat(os.Getenv("FUNDING_BOTTOM_RATE"), 64)
	crazyRate, _ := strconv.ParseFloat(os.Getenv("FUNDING_CRAZY_RATE"), 64)
	fixedAmount, _ := strconv.ParseFloat(os.Getenv("FUNDING_FIXED_AMOUNT"), 64)
	increaseRate, _ := strconv.ParseFloat(os.Getenv("FUNDING_INCREASE_RATE"), 64)
	telegramId, _ := strconv.ParseInt(os.Getenv("TELEGRAM_MANAGE_ID"), 10, 64)
	invalidRate, _ := strconv.ParseFloat(os.Getenv("INVALID_RATE"), 64)
	submitOffer := os.Getenv("AUTO_SUBMIT_FUNDING") == "Y"
	dayRange := strings.Split(os.Getenv("CRAZY_DAY_RANGE"), ",")
	minDay, _ := strconv.Atoi(os.Getenv("OFFICIAL_MIN_DAY"))
	maxDay, _ := strconv.Atoi(os.Getenv("OFFICIAL_MAX_DAY"))
	reLendTime, _ := strconv.ParseInt(os.Getenv("AUTO_CANCEL_RELEND_TIME"), 10, 64)
	crazyDayRange := make(map[int]float64, 0)

	if len(dayRange) != 0 && len(dayRange)%2 == 0 {
		for i := 0; i < len(dayRange); i += 2 {
			day, _ := strconv.Atoi(dayRange[i])
			rate, _ := strconv.ParseFloat(dayRange[i+1], 64)

			if day < minDay || day > maxDay {
				continue
			}
			crazyDayRange[day] = rate
		}
	}

	return &ConfigManage{
		RWMutex:       sync.RWMutex{},
		BottomRate:    bottomRate,
		FixedAmount:   fixedAmount,
		Day:           minDay,
		CrazyRate:     crazyRate,
		IncreaseRate:  increaseRate,
		TelegramId:    telegramId,
		SubmitOffer:   submitOffer,
		InvalidRate:   invalidRate,
		CrazyDayRange: crazyDayRange,
		Weights: map[string]int{
			"book01":   1,
			"book02":   1,
			"book03":   10,
			"avg100":   1,
			"avg10000": 3,
		},

		OfficialMaxDay: maxDay,
		OfficialMinDay: minDay,
		AutoCancelTime: reLendTime,
		NotifyRate:     0.001,
	}
}

func (config *ConfigManage) ConvertToMap() map[string]interface{} {
	return structs.Map(config)
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

func (config *ConfigManage) GetInvalidRate() float64 {
	config.Lock()
	defer config.Unlock()
	return config.InvalidRate
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

func (config *ConfigManage) GetCrazyDayRange() map[int]float64 {
	config.Lock()
	defer config.Unlock()
	return config.CrazyDayRange
}

func (config *ConfigManage) GetDayByRate(rate float64) int {
	config.Lock()
	defer config.Unlock()
	day := config.Day

	for days, crazyRate := range config.CrazyDayRange {
		if rate >= crazyRate && day <= days {
			day = days
		}
	}
	return day
}

func (config *ConfigManage) GetAutoCancelTime() int64 {
	config.Lock()
	defer config.Unlock()
	return config.AutoCancelTime
}

func (config *ConfigManage) GetNotifyRate() float64 {
	config.Lock()
	defer config.Unlock()
	return config.NotifyRate
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
	config.InvalidRate = rate
}

func (config *ConfigManage) SetWeights(key string, increment int) {
	config.Lock()
	defer config.Unlock()
	val, ok := config.Weights[key]
	if ok && (val+increment) > 0 {
		config.Weights[key] += increment
	}
}

func (config *ConfigManage) SetCrazyDayRange(rateMapDay string) {
	config.Lock()
	defer config.Unlock()

	dayRange := strings.Split(rateMapDay, ",")

	crazyDayRange := make(map[int]float64, 0)
	if len(dayRange) != 0 && len(dayRange)%2 == 0 {
		for i := 0; i < len(dayRange); i += 2 {
			day, _ := strconv.Atoi(dayRange[i])
			rate, _ := strconv.ParseFloat(dayRange[i+1], 64)

			if day < config.OfficialMinDay || day > config.OfficialMaxDay {
				continue
			}

			crazyDayRange[day] = rate
		}
	}
	config.CrazyDayRange = crazyDayRange
}

func (config *ConfigManage) SetAutoCancelTime(cancelTime int64) {
	config.Lock()
	defer config.Unlock()
	if cancelTime >= 10 {
		config.AutoCancelTime = cancelTime
	}
}

func (config *ConfigManage) SetNotifyRate(notifyRate float64) {
	config.Lock()
	defer config.Unlock()

	if notifyRate <= 0 {
		return
	}
	config.NotifyRate = notifyRate
}

// 權重初始化
func (config *ConfigManage) WeightsInit() {
	config.Lock()
	defer config.Unlock()
	config.Weights = map[string]int{
		"book01":   1,
		"book02":   1,
		"book03":   10,
		"avg100":   1,
		"avg10000": 3,
	}
}
