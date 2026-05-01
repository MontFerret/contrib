package dom

import (
	"context"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type domEventEvaluator interface {
	ContextID() cdpruntime.ExecutionContextID
	Eval(ctx context.Context, fn *eval.Function) error
}

func subscribeDOMTargetEvents(
	ctx context.Context,
	api domBindingRuntime,
	evaluator domEventEvaluator,
	targetID cdpruntime.RemoteObjectID,
	subscription runtime.Subscription,
) (runtime.Stream, error) {
	options, err := parseDOMEventOptions(ctx, subscription.Options)

	if err != nil {
		return nil, err
	}

	config := buildDOMEventTemplateOptions(options)

	return subscribeDOMEvents(
		ctx,
		api,
		evaluator.ContextID(),
		func(ctx context.Context, bindingName string) error {
			return evaluator.Eval(
				ctx,
				templates.AddEventListener(targetID, subscription.EventName, bindingName, config),
			)
		},
		func(ctx context.Context, bindingName string) error {
			return evaluator.Eval(
				ctx,
				templates.RemoveEventListener(targetID, bindingName),
			)
		},
	)
}
