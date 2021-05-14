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
		UpdateRatesWorkflow *UpdateRatesWorkflow
		ExchangeActivities  *ExchangeActivities
	}

	CronWorker struct {
		deps workerDeps
	}
)

func CreateWorker(deps workerDeps) error {
	var cronWorker = new(LazyWorker)

	deps.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) (startErr error) {
			deps.Logger.Info(ctx, "Registering Temporal cron workers")
			if deps.LazyTemporalClient == nil {
				return fmt.Errorf("temporal client wasn't created")
			}
			queueName := deps.Config.Get(queueNameKey).String()
			maxConcurrentWorkers := deps.Config.Get(maxConcurrentWorkersKey).Int()

			worker := worker.New(deps.LazyTemporalClient.Client, queueName, worker.Options{
				MaxConcurrentActivityExecutionSize:     maxConcurrentWorkers,
				MaxConcurrentWorkflowTaskExecutionSize: maxConcurrentWorkers,
			})
			worker.RegisterWorkflow(deps.UpdateRatesWorkflow.UpdateRates)
			worker.RegisterActivity(deps.ExchangeActivities.GetRates)
			worker.RegisterActivity(deps.ExchangeActivities.UpdateRates)

			if startErr = worker.Start(); startErr != nil {
				return
			}
			cronWorker.workers = append(cronWorker.workers, worker)
			return
		},
		OnStop: func(ctx context.Context) error {
			for _, registeredWorker := range cronWorker.workers {
				if registeredWorker != nil {
					registeredWorker.Stop()
				}
			}
			return nil
		},
	})
	return nil
}
