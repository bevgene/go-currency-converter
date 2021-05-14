package mortar

import (
	"github.com/bevgene/go-currency-rate/app/clients"
	"github.com/bevgene/go-currency-rate/app/temporal"
	"go.uber.org/fx"
)

func TemporalFxOptions() fx.Option {
	return fx.Options(
		fx.Provide(
			clients.CreateTemporalClient,
		),
		fx.Invoke(temporal.CreateWorker),
		fx.Invoke(temporal.CreateCronStarter),
		fx.Provide(
			temporal.CreateUpdateRatesWorkflow,
			temporal.CreateActivities,
		),
	)
}
