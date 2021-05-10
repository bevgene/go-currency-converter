package model

import "time"

type ExchangeRatesModel struct {
	Success   bool               `json:"success"`
	Date      string             `json:"date"`
	Base      string             `json:"base"`
	Timestamp int64              `json:"timestamp"`
	Rates     map[string]float64 `json:"rates"`
}

type ExchangeRateDocument struct {
	Rates     map[string]float64 `bson:"rates"`
	CreatedAt time.Time          `bson:"created_at"`
}

func ConvertExchangeRatesModel(model ExchangeRatesModel) (result ExchangeRateDocument) {
	result.Rates = model.Rates
	result.CreatedAt = time.Unix(model.Timestamp, 0)
	return
}
