package model

type DailyInterestReport struct {
	Balance       float64                  `json:"錢包總額"`
	TotalInterest float64                  `json:"利息總額"`
	InterestList  []map[string]interface{} `json:"利息清單"`
}
