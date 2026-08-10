package dom

import (
	"context"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type elementExecutor struct {
	exec       *eval.Runtime
	document   *HTMLDocument
	generation uint64
}

func newElementExecutor(exec *eval.Runtime, document *HTMLDocument, generation uint64) *elementExecutor {
	return &elementExecutor{
		exec:       exec,
		document:   document,
		generation: generation,
	}
}

func (exec *elementExecutor) ensureAttached() error {
	if exec.document == nil {
		return nil
	}

	if !exec.document.isGenerationCurrent(exec.generation) {
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

	if refreshErr := exec.document.refresh(ctx, exec.generation); refreshErr != nil {
		return refreshErr
	}

	return drivers.ErrDetached
}

func (exec *elementExecutor) Eval(ctx context.Context, fn *eval.Function) error {
	if err := exec.ensureAttached(); err != nil {
		return err
	}

	return exec.normalizeError(ctx, exec.exec.Eval(ctx, fn))
}

func (exec *elementExecutor) EvalValue(ctx context.Context, fn *eval.Function) (runtime.Value, error) {
	if err := exec.ensureAttached(); err != nil {
		return runtime.None, err
	}

	value, err := exec.exec.EvalValue(ctx, fn)

	return value, exec.normalizeError(ctx, err)
}

func (exec *elementExecutor) EvalResult(ctx context.Context, fn *eval.Function) (runtime.Value, error) {
	if err := exec.ensureAttached(); err != nil {
		return runtime.None, err
	}

	value, err := exec.exec.EvalResult(ctx, fn)

	return value, exec.normalizeError(ctx, err)
}

func (exec *elementExecutor) EvalList(ctx context.Context, fn *eval.Function) (*runtime.Array, error) {
	if err := exec.ensureAttached(); err != nil {
		return runtime.EmptyArray(), err
	}

	value, err := exec.exec.EvalList(ctx, fn)

	return value, exec.normalizeError(ctx, err)
}

func (exec *elementExecutor) EvalElement(ctx context.Context, fn *eval.Function) (runtime.Value, error) {
	if err := exec.ensureAttached(); err != nil {
		return runtime.None, err
	}

	value, err := exec.exec.EvalElement(ctx, fn)

	return value, exec.normalizeError(ctx, err)
}

func (exec *elementExecutor) EvalElements(ctx context.Context, fn *eval.Function) (*runtime.Array, error) {
	if err := exec.ensureAttached(); err != nil {
		return runtime.EmptyArray(), err
	}

	value, err := exec.exec.EvalElements(ctx, fn)

	return value, exec.normalizeError(ctx, err)
}

func (exec *elementExecutor) Runtime() *eval.Runtime {
	return exec.exec
}

func (exec *elementExecutor) ContextID() cdpruntime.ExecutionContextID {
	return exec.exec.ContextID()
}
