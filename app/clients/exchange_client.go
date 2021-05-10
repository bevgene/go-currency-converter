package clients

import (
	"context"
	"fmt"
	"github.com/bevgene/go-currency-rate/app/model"
	"github.com/go-masonry/mortar/interfaces/cfg"
	"github.com/go-masonry/mortar/interfaces/log"
	"go.uber.org/fx"
	"net/http"
)

//go:generate mockgen -source=exchange_client.go -destination=mock/exchange_client_mock.go

type (
	ExchangeClient interface {
		GetRate(context.Context, string, string) (float32, error)
		GetRates(context.Context) (model.ExchangeRatesModel, error)
	}

	exchangeClientImplDeps struct {
		fx.In

		Logger    log.Logger
		Config    cfg.Config
		Lifecycle fx.Lifecycle
	}

	exchangeClientImpl struct {
		deps   exchangeClientImplDeps
		client *http.Client
		url    string
	}
)

const (
	exchangeAPIKeyKey = "exchangerate.exchange.apiKey"
	exchangeUrlKey    = "exchangerate.exchange.url"
)

func CreateExchangeClient(deps exchangeClientImplDeps) (result ExchangeClient, err error) {
	url := deps.Config.Get(exchangeUrlKey).String()
	apiKey := deps.Config.Get(exchangeAPIKeyKey).String()

	authUrl := fmt.Sprintf("%s?access_key=%s", url, apiKey)
	deps.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) (err error) {
			var req *http.Request
			if req, err = http.NewRequest("GET", authUrl, nil); err != nil {
				return
			}
			_, err = http.DefaultClient.Do(req)
			return
		},
	})
	result = &exchangeClientImpl{
		deps:   deps,
		client: &http.Client{},
		url:    authUrl,
	}
	return
}

func (impl *exchangeClientImpl) GetRate(ctx context.Context, currencyFrom string, currencyTo string) (result float32, err error) {
	panic("implement me")
}

func (impl *exchangeClientImpl) GetRates(ctx context.Context) (model.ExchangeRatesModel, error) {
	panic("implement me")
}
