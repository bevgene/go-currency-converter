package temporal

import (
	"github.com/bevgene/go-currency-rate/app/model"
	"github.com/go-masonry/mortar/interfaces/cfg"
	"github.com/go-masonry/mortar/interfaces/log"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"
	"time"
)

type (
	updateRatesWorkflowDeps struct {
		fx.In

		Config             cfg.Config
		Logger             log.Logger
		ExchangeActivities *ExchangeActivities
	}

	UpdateRatesWorkflow struct {
		deps updateRatesWorkflowDeps
	}
)

func CreateUpdateRatesWorkflow(deps updateRatesWorkflowDeps) *UpdateRatesWorkflow {
	return &UpdateRatesWorkflow{
		deps: deps,
	}
}

func (impl *UpdateRatesWorkflow) UpdateRates(ctx workflow.Context) (err error) {
	workflow.GetLogger(ctx).Info("Cron workflow started.", "StartTime", workflow.Now(ctx))
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout:    time.Minute,
		ScheduleToStartTimeout: time.Minute,
	}
	ctx1 := workflow.WithActivityOptions(ctx, activityOptions)
	var rates model.ExchangeRatesModel
	if err = workflow.ExecuteActivity(ctx1, impl.deps.ExchangeActivities.GetRates).Get(ctx, &rates); err != nil {
		workflow.GetLogger(ctx).Error("Cron job failed.", "Error", err)
		return
	}

	document := model.ConvertExchangeRatesModel(rates)

	err = workflow.ExecuteActivity(ctx1, impl.deps.ExchangeActivities.UpdateRates, document).Get(ctx, nil)
	if err != nil {
		workflow.GetLogger(ctx).Error("Cron job failed.", "Error", err)

	}
	workflow.GetLogger(ctx).Info("Cron workflow finished.", "FinishTime", workflow.Now(ctx))
	return
}
