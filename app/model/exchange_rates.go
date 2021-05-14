package model

import "time"

type ExchangeRatesModel struct {
	Success   bool               `json:"success"`
	Date      string             `json:"date"`
	Base      string             `json:"base"`
	Timestamp int64              `json:"timestamp"`
	Rates     map[string]float32 `json:"rates"`
}

type ExchangeRateDocument struct {
	Base      string             `bson:"base"`
	Rates     map[string]float32 `bson:"rates"`
	CreatedAt time.Time          `bson:"created_at"`
}

func ConvertExchangeRatesModel(model ExchangeRatesModel) (result *ExchangeRateDocument) {
	result = &ExchangeRateDocument{
		Base:      model.Base,
		Rates:     model.Rates,
		CreatedAt: time.Unix(model.Timestamp, 0),
	}
	return
}
