package tests

import (
	"context"
	currencyconverter "github.com/bevgene/go-currency-rate/api"
	"github.com/bevgene/go-currency-rate/app/clients"
	mock_clients "github.com/bevgene/go-currency-rate/app/clients/mock"
	"github.com/go-masonry/mortar/interfaces/cfg"
	confkeys "github.com/go-masonry/mortar/interfaces/cfg/keys"
	"github.com/go-masonry/mortar/interfaces/http/client"
	"github.com/go-masonry/mortar/utils"
	"github.com/golang/mock/gomock"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"net"
	"net/http"
	"net/url"
	"testing"
)

type (
	CurrencyConverterClient interface {
		currencyconverter.CurrencyConverterClient
	}

	currencyConverterClientImplDeps struct {
		fx.In

		Config            cfg.Config
		HTTPClientBuilder client.NewHTTPClientBuilder
	}

	currencyConverterClientImpl struct {
		deps   currencyConverterClientImplDeps
		client utils.ProtobufHTTPClient
	}
)

const (
	convertPath = "/v1/convert"
)

func NewMockController(t *testing.T) (*gomock.Controller, context.Context) {
	return gomock.WithContext(context.Background(), t)
}

func CreateExchangeClientMock(mock *mock_clients.MockExchangeClient) clients.ExchangeClient {
	return mock
}

func CreateMongoClientMock(mock *mock_clients.MockMongoClient) clients.MongoClient {
	return mock
}

func CreateCurrencyConverterClient(deps currencyConverterClientImplDeps) CurrencyConverterClient {
	httpClient := deps.HTTPClientBuilder().Build()
	return &currencyConverterClientImpl{
		deps:   deps,
		client: utils.CreateProtobufHTTPClient(httpClient, convertHTTPStatusCodeToGRPCError, nil),
	}
}

func (impl *currencyConverterClientImpl) Convert(ctx context.Context, request *currencyconverter.ConvertRequest, opts ...grpc.CallOption) (result *currencyconverter.ConvertResponse, err error) {
	err = impl.callCurrencyConverter(ctx, convertPath, request, &result)
	return
}

func (impl *currencyConverterClientImpl) callCurrencyConverter(ctx context.Context, path string, request proto.Message, response interface{}) (err error) {
	serverPort := impl.deps.Config.Get(confkeys.ExternalRESTPort).String()
	endpointURL := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort("localhost", serverPort),
		Path:   path,
	}
	return impl.client.Do(ctx, http.MethodPost, endpointURL.String(), request, response)
}

func convertHTTPStatusCodeToGRPCError(httpStatus int) (result *status.Status) {
	if httpStatus > http.StatusAccepted {
		result = status.Newf(codes.Internal, "service returned http status code: %d", httpStatus)
	}
	return
}
