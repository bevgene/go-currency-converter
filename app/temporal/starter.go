package temporal

import (
	"context"
	"fmt"
	"github.com/bevgene/go-currency-rate/app/clients"
	"github.com/go-masonry/mortar/interfaces/cfg"
	"github.com/go-masonry/mortar/interfaces/log"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
)

type (
	cronStarterDeps struct {
		fx.In

		Lifecycle           fx.Lifecycle
		Config              cfg.Config
		Logger              log.Logger
		TemporalClient      *clients.LazyClient
		UpdateRatesWorkflow *UpdateRatesWorkflow
	}

	CronStarter struct {
		deps       cronStarterDeps
		WorkflowID string
		RunID      string
	}
)

func CreateCronStarter(deps cronStarterDeps) error {
	cronStarter := &CronStarter{
		deps: deps,
	}

	deps.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) (err error) {
			workflowName := deps.Config.Get(workflowNameKey).String()
			workflowID := fmt.Sprintf("cron_%s", workflowName)
			queueName := deps.Config.Get(queueNameKey).String()
			cronSchedule := deps.Config.Get(cronScheduleKey).String()
			workflowOptions := client.StartWorkflowOptions{
				ID:           workflowID,
				TaskQueue:    queueName,
				CronSchedule: cronSchedule,
			}
			var workflowRun client.WorkflowRun
			if workflowRun, err = cronStarter.deps.TemporalClient.Client.ExecuteWorkflow(context.Background(), workflowOptions, deps.UpdateRatesWorkflow.UpdateRates); err != nil {
				deps.Logger.WithError(err).Error(ctx, "failed workflow execution")
				return
			}
			cronStarter.WorkflowID = workflowRun.GetID()
			cronStarter.RunID = workflowRun.GetRunID()
			deps.Logger.WithField("workflow id", cronStarter.WorkflowID).WithField("run_id", cronStarter.RunID).Info(ctx, "starter started workflow")
			return
		},
		OnStop: func(ctx context.Context) (err error) {
			if len(cronStarter.WorkflowID) > 0 && len(cronStarter.RunID) > 0 {
				err = cronStarter.deps.TemporalClient.Client.TerminateWorkflow(ctx, cronStarter.WorkflowID, cronStarter.RunID, "stopping")
			}
			return
		},
	})
	return nil
}
