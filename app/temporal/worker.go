package temporal

import (
	"context"
	"fmt"
	"github.com/bevgene/go-currency-rate/app/clients"
	"github.com/go-masonry/mortar/interfaces/cfg"
	"github.com/go-masonry/mortar/interfaces/log"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

type (
	LazyWorker struct {
		workers []worker.Worker
	}
	workerDeps struct {
		fx.In

		LazyTemporalClient  *clients.LazyClient
		Config              cfg.Config
		Logger              log.Logger
		Lifecycle           fx.Lifecycle
		UpdateRatesWorkflow UpdateRatesWorkflow
		ExchangeActivities  ExchangeActivities
	}
)

const (
	queueNameKey            = "exchangerate.temporal.queue"
	maxConcurrentWorkersKey = "exchangerate.temporal.maxConcurrentWorkers"
)

func CreateWorker(deps workerDeps) error {
	var cronWorker worker.Worker

	deps.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			deps.Logger.Info(ctx, "Registering Temporal cron workers")
			if deps.LazyTemporalClient == nil {
				return fmt.Errorf("temporal client wasn't created")
			}
			queueName := deps.Config.Get(queueNameKey).String()
			maxConcurrentWorkers := deps.Config.Get(maxConcurrentWorkersKey).Int()

			cronWorker = worker.New(deps.LazyTemporalClient, queueName, worker.Options{
				MaxConcurrentActivityExecutionSize:     maxConcurrentWorkers,
				MaxConcurrentWorkflowTaskExecutionSize: maxConcurrentWorkers,
			})
			cronWorker.RegisterWorkflow(deps.UpdateRatesWorkflow)
			cronWorker.RegisterActivity(deps.ExchangeActivities)

			err := cronWorker.Start()
			return err
		},
		OnStop: func(ctx context.Context) error {
			cronWorker.Stop()
			return nil
		},
	})
	return nil
}
