package tests

import (
	"context"
	"fmt"
	currencyconverter "github.com/bevgene/go-currency-rate/api"
	mock_clients "github.com/bevgene/go-currency-rate/app/clients/mock"
	"github.com/bevgene/go-currency-rate/app/model"
	"github.com/bevgene/go-currency-rate/app/mortar"
	"github.com/go-masonry/mortar/interfaces/log"
	"github.com/go-masonry/mortar/providers"
	"github.com/golang/mock/gomock"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"testing"
)

type (
	componentTestSuiteDeps struct {
		fx.In

		ServiceClient   CurrencyConverterClient
		MockCtrl        *gomock.Controller
		MockMongoClient *mock_clients.MockMongoClient
		Ctx             context.Context
		Logger          log.Logger
		ExpectedRates   *model.ExchangeRateDocument
	}

	componentTestSuite struct {
		suite.Suite

		TestApp *fxtest.App
		deps    componentTestSuiteDeps
	}
)

func TestComponent(t *testing.T) {
	suite.Run(t, new(componentTestSuite))
}

func (impl *componentTestSuite) SetupTest() {
	testApp := fxtest.New(
		impl.T(),
		fx.Supply(impl.T()),
		mortar.ViperFxOption("../config/config.yml", "../config/config_test.yml"),
		mortar.LoggerFxOption(),
		mortar.HttpServerFxOptions(),
		mortar.HttpClientFxOptions(),
		mortar.InternalHttpHandlersFxOptions(),
		mortar.ServiceAPIsAndOtherDependenciesFxOption(),
		fx.Provide(
			NewMockController,
			CreateCurrencyConverterClient,
			mock_clients.NewMockMongoClient,
			mock_clients.NewMockExchangeClient,
			CreateExchangeClientMock,
			CreateMongoClientMock,
			CreateLazyMongoClient,
			GetRatesDocument,
		),
		providers.BuildMortarWebServiceFxOption(),
		fx.Populate(&impl.deps),
	)
	impl.TestApp = testApp
	impl.TestApp.RequireStart()
}

func (impl *componentTestSuite) TearDownTest() {
	if impl.deps.MockCtrl != nil {
		impl.deps.MockCtrl.Finish()
	}

	if impl.TestApp != nil {
		impl.TestApp.RequireStop()
	}
}

func (impl *componentTestSuite) TestConvert() {
	t := impl.T()

	params := gopter.DefaultTestParametersWithSeed(gopterSeed)
	props := gopter.NewProperties(params)
	props.Property("happy convert test", impl.happyConvert(t))
	props.TestingRun(t)
}

func (impl *componentTestSuite) happyConvert(t *testing.T) gopter.Prop {
	return prop.ForAll(
		func(request *currencyconverter.ConvertRequest) bool {
			var response *currencyconverter.ConvertResponse
			var err error
			impl.deps.MockMongoClient.EXPECT().GetLatestRateDocument(gomock.Any()).Return(impl.deps.ExpectedRates, nil)
			response, err = impl.deps.ServiceClient.Convert(impl.deps.Ctx, request)
			var ok bool
			ok = assert.NoError(t, err, "failed to retrieve convert response")
			if !ok {
				return ok
			}
			var expectedAmount float32
			expectedAmount, err = impl.calculateExpectedAmount(request.GetCurrencyFrom(), request.GetCurrencyTo(),
				request.GetAmountFrom())
			ok = assert.NoError(t, err, "failed to calculate expected amount")
			if !ok {
				return ok
			}
			return assert.NotNil(t, response, "response should not be empty") &&
				assert.EqualValues(t, request.GetCurrencyTo(), response.GetCurrency()) &&
				assert.Equal(t, expectedAmount, response.GetAmount())
		},
		ConvertRequestGenerator(),
	)
}

func (impl *componentTestSuite) calculateExpectedAmount(currencyFrom, currencyTo string, amountFrom float32) (result float32, err error) {
	var rateFrom, rateTo float32
	if rateFrom, err = impl.getRate(currencyFrom); err != nil {
		return
	}
	if rateTo, err = impl.getRate(currencyTo); err != nil {
		return
	}
	if rateFrom <= 0 {
		err = fmt.Errorf("illegal rate [%f] for %s", rateFrom, currencyFrom)
		return
	}
	result = amountFrom * rateTo / rateFrom
	return
}

func (impl *componentTestSuite) getRate(currency string) (result float32, err error) {
	var ok bool
	if result, ok = impl.deps.ExpectedRates.Rates[currency]; !ok {
		err = fmt.Errorf("unknown currency %s", currency)
		return
	}
	return
}
