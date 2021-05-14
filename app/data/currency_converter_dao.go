package data

import (
	"context"
	"github.com/bevgene/go-currency-rate/app/clients"
	"github.com/bevgene/go-currency-rate/app/model"
	"github.com/go-masonry/mortar/interfaces/cfg"
	"github.com/go-masonry/mortar/interfaces/log"
	"go.uber.org/fx"
)

type (
	CurrencyRateDao interface {
		GetRates(ctx context.Context) (*model.ExchangeRateDocument, error)
	}

	currencyRateDaoImplDeps struct {
		fx.In

		Logger          log.Logger
		Config          cfg.Config
		Lifecycle       fx.Lifecycle
		LazyMongoClient *clients.LazyMongoClient
	}

	currencyRateDaoImpl struct {
		deps currencyRateDaoImplDeps
	}
)

func CreateCurrencyRateDao(deps currencyRateDaoImplDeps) (result CurrencyRateDao) {
	result = &currencyRateDaoImpl{
		deps: deps,
	}
	return
}

func (impl *currencyRateDaoImpl) GetRates(ctx context.Context) (*model.ExchangeRateDocument, error) {
	return impl.deps.LazyMongoClient.Client.GetLatestRateDocument(ctx)
}
