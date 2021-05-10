package temporal

import (
	"context"
	"fmt"
	"github.com/bevgene/go-currency-rate/app/clients"
	"github.com/go-masonry/mortar/interfaces/cfg"
	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
)

type (
	cronStarterDeps struct {
		fx.In

		Lifecycle           fx.Lifecycle
		Config              cfg.Config
		TemporalClient      clients.LazyClient
		UpdateRatesWorkflow UpdateRatesWorkflow
	}

	CronStarter struct {
		deps       cronStarterDeps
		WorkflowID string
		RunID      string
	}
)

func CreateCronStarter(deps cronStarterDeps) (CronStarter, error) {
	cronStarter := CronStarter{
		deps: deps,
	}

	deps.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) (err error) {
			workflowID := fmt.Sprintf("cron_%s", uuid.New())
			workflowOptions := client.StartWorkflowOptions{
				ID:        workflowID,
				TaskQueue: "cron",
				// The cron spec is as following:
				// ┌───────────── minute (0 - 59)
				// │ ┌───────────── hour (0 - 23)
				// │ │ ┌───────────── day of the month (1 - 31)
				// │ │ │ ┌───────────── month (1 - 12)
				// │ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday)
				// │ │ │ │ │
				// │ │ │ │ │
				// * * * * *

				CronSchedule: "* * * * *", // every 1 minute
			}
			var workflowRun client.WorkflowRun
			if workflowRun, err = cronStarter.deps.TemporalClient.Client.ExecuteWorkflow(context.Background(), workflowOptions, deps.UpdateRatesWorkflow.UpdateRates); err != nil {
				return
			}
			cronStarter.WorkflowID = workflowRun.GetID()
			cronStarter.RunID = workflowRun.GetRunID()
			return
		},
		OnStop: func(ctx context.Context) (err error) {
			if len(cronStarter.WorkflowID) > 0 && len(cronStarter.RunID) > 0 {
				err = cronStarter.deps.TemporalClient.Client.TerminateWorkflow(ctx, cronStarter.WorkflowID, cronStarter.RunID, "stopping")
			}
			return
		},
	})
	return cronStarter, nil
}
