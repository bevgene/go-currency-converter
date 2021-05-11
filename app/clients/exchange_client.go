package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bevgene/go-currency-rate/app/model"
	"github.com/go-masonry/mortar/interfaces/cfg"
	"github.com/go-masonry/mortar/interfaces/http/client"
	"github.com/go-masonry/mortar/interfaces/log"
	"go.uber.org/fx"
	"io/ioutil"
	"net/http"
)

//go:generate mockgen -source=exchange_client.go -destination=mock/exchange_client_mock.go

type (
	ExchangeClient interface {
		GetRates(context.Context) (*model.ExchangeRatesModel, error)
	}

	exchangeClientImplDeps struct {
		fx.In

		Logger            log.Logger
		Config            cfg.Config
		Lifecycle         fx.Lifecycle
		HTTPClientBuilder client.NewHTTPClientBuilder
	}

	exchangeClientImpl struct {
		deps   exchangeClientImplDeps
		client *http.Client
		url    string
	}
)

const (
	exchangeAPIKeyKey  = "exchangerate.exchange.apiKey"
	exchangeUrlKey     = "exchangerate.exchange.url"
	exchangeTimeoutKey = "exchangerate.exchange.timeout"
)

func CreateExchangeClient(deps exchangeClientImplDeps) (result ExchangeClient, err error) {
	url := deps.Config.Get(exchangeUrlKey).String()
	apiKey := deps.Config.Get(exchangeAPIKeyKey).String()
	timeout := deps.Config.Get(exchangeTimeoutKey).Duration()

	authUrl := fmt.Sprintf("%s?access_key=%s", url, apiKey)
	httpClient := deps.HTTPClientBuilder().WithPreconfiguredClient(&http.Client{Timeout: timeout}).Build()
	deps.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) (err error) {
			var req *http.Request
			if req, err = http.NewRequest("GET", authUrl, nil); err != nil {
				return
			}
			_, err = httpClient.Do(req)
			return
		},
	})
	result = &exchangeClientImpl{
		deps:   deps,
		client: httpClient,
		url:    authUrl,
	}
	return
}

func (impl *exchangeClientImpl) GetRates(ctx context.Context) (result *model.ExchangeRatesModel, err error) {
	var req *http.Request
	if req, err = http.NewRequest("GET", impl.url, nil); err != nil {
		impl.deps.Logger.WithError(err).Error(ctx, "failed to create a new request")
		return
	}

	var res *http.Response
	if res, err = impl.client.Do(req); err != nil {
		impl.deps.Logger.WithError(err).Error(ctx, "failed to execute request")
		return
	}

	defer func() {
		var closeErr error
		if closeErr = res.Body.Close(); closeErr != nil {
			impl.deps.Logger.WithError(closeErr).Error(ctx, "failed to close response body")
		}
	}()
	var body []byte
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		impl.deps.Logger.WithError(err).Error(ctx, "failed to read response body")
		return
	}
	var parsedRates model.ExchangeRatesModel
	if err = json.Unmarshal(body, &parsedRates); err != nil {
		impl.deps.Logger.WithError(err).Error(ctx, "failed to parse response body")
		return
	}
	result = &parsedRates
	return
}
