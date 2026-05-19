package dom

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type elementEvaluator interface {
	Eval(ctx context.Context, fn *eval.Function) error
	EvalValue(ctx context.Context, fn *eval.Function) (runtime.Value, error)
}
