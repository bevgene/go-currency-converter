package controllers

import (
	"context"
	"github.com/bevgene/go-currency-rate/app/data"
	"github.com/go-masonry/mortar/interfaces/cfg"

	currencyconverter "github.com/bevgene/go-currency-rate/api"

	"github.com/go-masonry/mortar/interfaces/log"
	"go.uber.org/fx"
)

type (
	CurrencyRateController interface {
		currencyconverter.CurrencyConverterServer
	}

	currencyRateControllerImplDeps struct {
		fx.In

		Logger          log.Logger
		Config          cfg.Config
		CurrencyRateDao data.CurrencyRateDao
	}

	currencyRateControllerImpl struct {
		*currencyconverter.UnimplementedCurrencyConverterServer
		deps currencyRateControllerImplDeps
	}
)

func CreateCurrencyRateController(deps currencyRateControllerImplDeps) CurrencyRateController {
	return &currencyRateControllerImpl{
		deps: deps,
	}
}

func (impl *currencyRateControllerImpl) Convert(ctx context.Context, request *currencyconverter.ConvertRequest) (*currencyconverter.ConvertResponse, error) {
	panic("implement me")
}
