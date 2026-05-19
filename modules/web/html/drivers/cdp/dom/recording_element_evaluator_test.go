package dom

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type recordingElementEvaluator struct {
	value          runtime.Value
	err            error
	evalCalls      int
	evalValueCalls int
}

func (rec *recordingElementEvaluator) Eval(_ context.Context, _ *eval.Function) error {
	rec.evalCalls++

	return rec.err
}

func (rec *recordingElementEvaluator) EvalValue(_ context.Context, _ *eval.Function) (runtime.Value, error) {
	rec.evalValueCalls++

	return rec.value, rec.err
}
