package temporal

import (
	"github.com/bevgene/go-currency-rate/app/model"
	"github.com/go-masonry/mortar/interfaces/cfg"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"
)

type (
	updateRatesWorkflowDeps struct {
		fx.In

		Config             cfg.Config
		ExchangeActivities ExchangeActivities
	}

	UpdateRatesWorkflow struct {
		deps updateRatesWorkflowDeps
	}
)

func CreateUpdateRatesWorkflow(deps updateRatesWorkflowDeps) UpdateRatesWorkflow {
	return UpdateRatesWorkflow{
		deps: deps,
	}
}

// UpdateRatesWorkflow executes on the given schedule
// The schedule is provided when starting the workflow
func (impl *UpdateRatesWorkflow) UpdateRates(ctx workflow.Context) (err error) {
	var rates model.ExchangeRatesModel
	if err = workflow.ExecuteActivity(ctx, impl.deps.ExchangeActivities.GetRates).Get(ctx, &rates); err != nil {
		return
	}

	document := model.ConvertExchangeRatesModel(rates)

	err = workflow.ExecuteActivity(ctx, impl.deps.ExchangeActivities.UpdateRates, &document).Get(ctx, nil)
	return
}
