package dom

import (
	"context"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type elementExecutor struct {
	document *HTMLDocument
	state    *documentState
}

func newElementExecutor(exec *eval.Runtime) *elementExecutor {
	return &elementExecutor{
		state: &documentState{eval: exec},
	}
}

func newStateElementExecutor(document *HTMLDocument, state *documentState) *elementExecutor {
	return &elementExecutor{
		document: document,
		state:    state,
	}
}

func (exec *elementExecutor) ensureAttached() error {
	if exec == nil {
		return drivers.ErrDetached
	}

	if exec.document == nil {
		return nil
	}

	if !exec.document.isCurrentState(exec.state) {
		return drivers.ErrDetached
	}

	return nil
}

func (exec *elementExecutor) normalizeError(ctx context.Context, err error) error {
	if err == nil || !eval.IsStaleError(err) {
		return err
	}

	if exec.document == nil {
		return err
	}

	if refreshErr := exec.document.refresh(ctx, exec.state); refreshErr != nil {
		return refreshErr
	}

	return drivers.ErrDetached
}

func (exec *elementExecutor) Eval(ctx context.Context, fn *eval.Function) error {
	return exec.run(ctx, func() error {
		return exec.state.eval.Eval(ctx, fn)
	})
}

func (exec *elementExecutor) EvalValue(ctx context.Context, fn *eval.Function) (runtime.Value, error) {
	return runElementResult(
		ctx,
		exec,
		func() runtime.Value { return runtime.None },
		func() (runtime.Value, error) { return exec.state.eval.EvalValue(ctx, fn) },
	)
}

func (exec *elementExecutor) EvalResult(ctx context.Context, fn *eval.Function) (runtime.Value, error) {
	return runElementResult(
		ctx,
		exec,
		func() runtime.Value { return runtime.None },
		func() (runtime.Value, error) { return exec.state.eval.EvalResult(ctx, fn) },
	)
}

func (exec *elementExecutor) EvalList(ctx context.Context, fn *eval.Function) (*runtime.Array, error) {
	return runElementResult(
		ctx,
		exec,
		runtime.EmptyArray,
		func() (*runtime.Array, error) { return exec.state.eval.EvalList(ctx, fn) },
	)
}

func (exec *elementExecutor) EvalElement(ctx context.Context, fn *eval.Function) (runtime.Value, error) {
	return runElementResult(
		ctx,
		exec,
		func() runtime.Value { return runtime.None },
		func() (runtime.Value, error) { return exec.state.eval.EvalElement(ctx, fn) },
	)
}

func (exec *elementExecutor) EvalElements(ctx context.Context, fn *eval.Function) (*runtime.Array, error) {
	return runElementResult(
		ctx,
		exec,
		runtime.EmptyArray,
		func() (*runtime.Array, error) { return exec.state.eval.EvalElements(ctx, fn) },
	)
}

func (exec *elementExecutor) run(ctx context.Context, operation func() error) error {
	if err := exec.ensureAttached(); err != nil {
		return err
	}

	return exec.normalizeError(ctx, operation())
}

func (exec *elementExecutor) ContextID() cdpruntime.ExecutionContextID {
	return exec.state.eval.ContextID()
}
