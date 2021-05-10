package validations

import (
	"context"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"

	currencyconverter "github.com/bevgene/go-currency-rate/api"
	"github.com/go-masonry/mortar/interfaces/log"
	"go.uber.org/fx"
)

type (
	CurrencyRateValidations interface {
		ValidateGetCurrencyRateRequest(ctx context.Context, request *currencyconverter.ConvertRequest) error
	}

	currencyRateValidationsImplDeps struct {
		fx.In
		Logger log.Logger
	}

	currencyRateValidationsImpl struct {
		deps currencyRateValidationsImplDeps
	}
)

func CreateCurrencyRateValidations(deps currencyRateValidationsImplDeps) CurrencyRateValidations {
	return &currencyRateValidationsImpl{
		deps: deps,
	}
}

func (impl *currencyRateValidationsImpl) ValidateGetCurrencyRateRequest(ctx context.Context, request *currencyconverter.ConvertRequest) (err error) {
	return combineErrors(
		impl.notEmpty(ctx, request.GetCurrencyTo(), "currencyTo"),
		impl.notEmpty(ctx, request.GetCurrencyFrom(), "currencyFrom"),
		func() error {
			if request.GetAmountFrom() < 0 {
				impl.deps.Logger.WithField("amount", request.GetAmountFrom()).Error(ctx, "amount cannot be negative")
				return status.Errorf(codes.InvalidArgument, "negative amount")
			}
			return nil
		}(),
	)
}

func combineErrors(errs ...error) (err error) {
	combinedErrors := multierr.Combine(errs...)
	if actualErrors := multierr.Errors(combinedErrors); len(actualErrors) > 0 {
		err = actualErrors[0]
	}
	return
}

func (impl *currencyRateValidationsImpl) notEmpty(ctx context.Context, value interface{}, fieldName string) (err error) {
	if value == reflect.Zero(reflect.TypeOf(value)).Interface() {
		impl.deps.Logger.WithField(fieldName, value).Error(ctx, "cannot be empty")
		err = status.Errorf(codes.InvalidArgument, "%s cannot be empty", fieldName)
	}

	return
}
