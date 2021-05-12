package services

import (
	"context"

	currencyconverter "github.com/bevgene/go-currency-rate/api"
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
		currencyconverter.UnimplementedCurrencyConverterServer
		deps currencyRateServiceImplDeps
	}
)

func CreateCurrencyRateService(deps currencyRateServiceImplDeps) currencyconverter.CurrencyConverterServer {
	return &currencyRateServiceImpl{
		deps: deps,
	}
}

func (impl *currencyRateServiceImpl) Convert(ctx context.Context, req *currencyconverter.ConvertRequest) (res *currencyconverter.ConvertResponse, err error) {
	if err = impl.deps.Validations.ValidateGetCurrencyRateRequest(ctx, req); err != nil {
		impl.deps.Logger.WithError(err).WithField("request", req).Error(ctx, "validation failed")
	}

	return impl.deps.Controller.Convert(ctx, req)
}
