package dom

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type (
	elementValueEvaluator interface {
		EvalValue(ctx context.Context, fn *eval.Function) (runtime.Value, error)
	}

	elementEvaluator interface {
		elementValueEvaluator
		Eval(ctx context.Context, fn *eval.Function) error
	}
)
