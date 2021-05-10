package controllers

import (
	"context"
	"github.com/bevgene/go-currency-rate/app/data"
	"github.com/go-masonry/mortar/interfaces/cfg"

	currencyrate "github.com/bevgene/go-currency-rate/api"

	"github.com/go-masonry/mortar/interfaces/log"
	"go.uber.org/fx"
)

type (
	CurrencyRateController interface {
		currencyrate.CurrencyRateFetcherServer
	}

	currencyRateControllerImplDeps struct {
		fx.In

		Logger          log.Logger
		Config          cfg.Config
		CurrencyRateDao data.CurrencyRateDao
	}

	currencyRateControllerImpl struct {
		*currencyrate.UnimplementedCurrencyRateFetcherServer
		deps currencyRateControllerImplDeps
	}
)

func CreateCurrencyRateController(deps currencyRateControllerImplDeps) CurrencyRateController {
	return &currencyRateControllerImpl{
		deps: deps,
	}
}

func (impl *currencyRateControllerImpl) GetCurrencyRate(ctx context.Context, request *currencyrate.GetCurrencyRateRequest) (*currencyrate.GetCurrencyRateResponse, error) {
	panic("implement me")
}
