package temporal

import (
	"context"
	"github.com/bevgene/go-currency-rate/app/clients"
	"github.com/bevgene/go-currency-rate/app/data"
	"github.com/bevgene/go-currency-rate/app/model"
	"go.uber.org/fx"
)

type (
	activityDeps struct {
		fx.In

		ExchangeClient  clients.ExchangeClient
		CurrencyRateDao data.CurrencyRateDao
	}

	ExchangeActivities struct {
		deps activityDeps
	}
)

func CreateActivities(deps activityDeps) ExchangeActivities {
	return ExchangeActivities{
		deps: deps,
	}
}

func (impl *ExchangeActivities) GetRates(ctx context.Context) (result model.ExchangeRatesModel, err error) {
	return impl.deps.ExchangeClient.GetRates(ctx)
}

func (impl *ExchangeActivities) UpdateRates(ctx context.Context, doc model.ExchangeRateDocument) error {
	return impl.deps.CurrencyRateDao.UpdateRates(ctx, doc)
}
