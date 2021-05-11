package controllers

import (
	"context"
	"fmt"
	"github.com/bevgene/go-currency-rate/app/data"
	"github.com/bevgene/go-currency-rate/app/model"
	"github.com/go-masonry/mortar/interfaces/cfg"
	"google.golang.org/protobuf/types/known/timestamppb"

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

func (impl *currencyRateControllerImpl) Convert(ctx context.Context, request *currencyconverter.ConvertRequest) (result *currencyconverter.ConvertResponse, err error) {
	var ratesDocument *model.ExchangeRateDocument
	if ratesDocument, err = impl.deps.CurrencyRateDao.GetRates(ctx); err != nil {
		impl.deps.Logger.WithError(err).WithField("request", request).Error(ctx, "failed fetching latest rates information from db")
		return
	}
	if ratesDocument == nil {
		err = fmt.Errorf("no information found in db")
		impl.deps.Logger.WithError(err).Error(ctx, "convert failed")
		return
	}
	currencyFrom := request.GetCurrencyFrom()
	currencyTo := request.GetCurrencyTo()
	amount := request.GetAmountFrom()
	var rateFrom, rateTo float32
	var ok bool

	if rateFrom, ok = ratesDocument.Rates[currencyFrom]; !ok {
		err = fmt.Errorf("unsupported currency %s", currencyFrom)
		impl.deps.Logger.WithError(err).WithField("request", request).Error(ctx, "convert failed")
		return
	}
	if rateTo, ok = ratesDocument.Rates[currencyTo]; !ok {
		err = fmt.Errorf("unsupported currency %s", currencyTo)
		impl.deps.Logger.WithError(err).WithField("request", request).Error(ctx, "convert failed")
	}

	result = &currencyconverter.ConvertResponse{
		Currency:        currencyTo,
		Amount:          0,
		CorrectnessTime: timestamppb.New(ratesDocument.CreatedAt),
	}

	if rateFrom > 0 {
		result.Amount = amount * rateTo / rateFrom
	}
	impl.deps.Logger.WithError(err).WithField("request", request).WithField("result", result).Info(ctx, "finished conversion")
	return
}
