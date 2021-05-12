package tests

import (
	"encoding/json"
	currencyconverter "github.com/bevgene/go-currency-rate/api"
	"github.com/bevgene/go-currency-rate/app/model"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"io/ioutil"
)

const (
	gopterSeed = 1620818024667484000
)

func ConvertRequestGenerator() gopter.Gen {
	return gopter.DeriveGen(
		func(currencyFrom string, currencyTo string, amount float32) *currencyconverter.ConvertRequest {
			return &currencyconverter.ConvertRequest{
				CurrencyFrom: currencyFrom,
				CurrencyTo:   currencyTo,
				AmountFrom:   amount,
			}
		},
		func(elem *currencyconverter.ConvertRequest) (string, string, float32) {
			return elem.GetCurrencyFrom(), elem.GetCurrencyTo(), elem.GetAmountFrom()
		},
		CurrencyGenerator(),
		CurrencyGenerator(),
		gen.Float32().SuchThat(func(f float32) bool { return f > 0 }),
	)
}

func CurrencyGenerator() gopter.Gen {
	return gen.OneConstOf("AED", "AFN", "ALL", "AMD", "ANG", "AOA", "ARS", "AUD", "AWG", "AZN", "BAM", "BBD", "BDT", "BGN", "BHD", "BIF", "BMD", "BND", "BOB", "BRL", "BSD", "BTC", "BTN", "BWP", "BYN", "BYR", "BZD", "CAD", "CDF", "CHF", "CLF", "CLP", "CNY", "COP", "CRC", "CUC", "CUP", "CVE", "CZK", "DJF", "DKK", "DOP", "DZD", "EGP", "ERN", "ETB", "EUR", "FJD", "FKP", "GBP", "GEL", "GGP", "GHS", "GIP", "GMD", "GNF", "GTQ", "GYD", "HKD", "HNL", "HRK", "HTG", "HUF", "IDR", "ILS", "IMP", "INR", "IQD", "IRR", "ISK", "JEP", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LVL", "LYD", "MAD", "MDL", "MGA", "MKD", "MMK", "MNT", "MOP", "MRO", "MUR", "MVR", "MWK", "MXN", "MYR", "MZN", "NAD", "NGN", "NIO", "NOK", "NPR", "NZD", "OMR", "PAB", "PEN", "PGK", "PHP", "PKR", "PLN", "PYG", "QAR", "RON", "RSD", "RUB", "RWF", "SAR", "SBD", "SCR", "SDG", "SEK", "SGD", "SHP", "SLL", "SOS", "SRD", "STD", "SVC", "SYP", "SZL", "THB", "TJS", "TMT", "TND", "TOP", "TRY", "TTD", "TWD", "TZS", "UAH", "UGX", "USD", "UYU", "UZS", "VEF", "VND", "VUV", "WST", "XAF", "XAG", "XAU", "XCD", "XDR", "XOF", "XPF", "YER", "ZAR", "ZMK", "ZMW", "ZWL")
}

func GetRatesDocument() (result *model.ExchangeRateDocument, err error) {
	var contentBytes []byte
	if contentBytes, err = ioutil.ReadFile("testdata/rates.json"); err != nil {
		return
	}
	var document model.ExchangeRateDocument
	if err = json.Unmarshal(contentBytes, &document); err != nil {
		return
	}
	result = &document
	return
}
