package mortar

import (
	"context"
	"github.com/bevgene/go-currency-rate/app/data"

	currencyrate "github.com/bevgene/go-currency-rate/api"
	"github.com/bevgene/go-currency-rate/app/controllers"
	"github.com/bevgene/go-currency-rate/app/services"
	"github.com/bevgene/go-currency-rate/app/validations"
	serverInt "github.com/go-masonry/mortar/interfaces/http/server"
	"github.com/go-masonry/mortar/providers/groups"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type workshopServiceDeps struct {
	fx.In

	// API Implementations, "Register" them as GRPCServiceAPI
	CurrencyRateFetcher currencyrate.CurrencyRateFetcherServer
}

func ServiceAPIsAndOtherDependenciesFxOption() fx.Option {
	return fx.Options(
		// GRPC Service APIs registration
		fx.Provide(fx.Annotated{
			Group:  groups.GRPCServerAPIs,
			Target: serviceGRPCServiceAPIs,
		}),
		// GRPC Gateway Generated Handlers registration
		fx.Provide(fx.Annotated{
			Group:  groups.GRPCGatewayGeneratedHandlers + ",flatten", // "flatten" does this [][]serverInt.GRPCGatewayGeneratedHandlers -> []serverInt.GRPCGatewayGeneratedHandlers
			Target: serviceGRPCGatewayHandlers,
		}),
		// All other tutorial dependencies
		serviceDependencies(),
	)
}

func serviceGRPCServiceAPIs(deps workshopServiceDeps) serverInt.GRPCServerAPI {
	return func(srv *grpc.Server) {
		currencyrate.RegisterCurrencyRateFetcherServer(srv, deps.CurrencyRateFetcher)
		// Any additional gRPC Implementations should be called here
	}
}

func serviceGRPCGatewayHandlers() []serverInt.GRPCGatewayGeneratedHandlers {
	return []serverInt.GRPCGatewayGeneratedHandlers{
		// Register service REST API
		func(mux *runtime.ServeMux, localhostEndpoint string) error {
			return currencyrate.RegisterCurrencyRateFetcherHandlerFromEndpoint(context.Background(), mux, localhostEndpoint, []grpc.DialOption{grpc.WithInsecure()})
		},
		// Any additional gRPC gateway registrations should be called here
	}
}

func serviceDependencies() fx.Option {
	return fx.Provide(
		services.CreateCurrencyRateService,
		controllers.CreateCurrencyRateController,
		validations.CreateCurrencyRateValidations,
		data.CreateCurrencyRateDao,
	)
}
