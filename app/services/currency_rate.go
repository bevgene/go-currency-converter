package services

import (
	"context"

	currencyrate "github.com/bevgene/go-currency-rate/api"
	"github.com/bevgene/go-currency-rate/app/controllers"
	"github.com/bevgene/go-currency-rate/app/validations"
	"github.com/go-masonry/mortar/interfaces/monitor"

	"github.com/go-masonry/mortar/interfaces/log"
	"go.uber.org/fx"
)

type (
	currencyRateServiceImplDeps struct {
		fx.In

		Logger      log.Logger
		Validations validations.CurrencyRateValidations
		Controller  controllers.CurrencyRateController
		Metrics     monitor.Metrics `optional:"true"`
	}

	currencyRateServiceImpl struct {
		currencyrate.UnimplementedCurrencyRateFetcherServer
		deps currencyRateServiceImplDeps
	}
)

func CreateCurrencyRateService(deps currencyRateServiceImplDeps) currencyrate.CurrencyRateFetcherServer {
	return &currencyRateServiceImpl{
		deps: deps,
	}
}

func (impl *currencyRateServiceImpl) GetCurrencyRate(ctx context.Context, req *currencyrate.GetCurrencyRateRequest) (res *currencyrate.GetCurrencyRateResponse, err error) {
	if err = impl.deps.Validations.ValidateGetCurrencyRateRequest(ctx, req); err != nil {
		impl.deps.Logger.WithError(err).WithField("request", req).Error(ctx, "validation failed")
	}

	return impl.deps.Controller.GetCurrencyRate(ctx, req)
}
