package eval

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/contrib/modules/html/drivers/common"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
	"github.com/rs/zerolog"
)

const (
	EmptyExecutionContextID = cdpruntime.ExecutionContextID(-1)
	EmptyObjectID           = cdpruntime.RemoteObjectID("")
)

type Runtime struct {
	logger    zerolog.Logger
	client    *cdp.Client
	frame     page.Frame
	contextID cdpruntime.ExecutionContextID
	resolver  *Resolver
}

func Create(
	ctx context.Context,
	logger zerolog.Logger,
	client *cdp.Client,
	frameID page.FrameID,
) (*Runtime, error) {
	world, err := client.Page.CreateIsolatedWorld(ctx, page.NewCreateIsolatedWorldArgs(frameID))

	if err != nil {
		return nil, err
	}

	return New(logger, client, frameID, world.ExecutionContextID), nil
}

func New(
	logger zerolog.Logger,
	client *cdp.Client,
	frameID page.FrameID,
	contextID cdpruntime.ExecutionContextID,
) *Runtime {
	rt := new(Runtime)
	rt.logger = common.
		LoggerWithName(logger.With(), "js-eval").
		Str("frame_id", string(frameID)).
		Int("context_id", int(contextID)).
		Logger()
	rt.client = client
	rt.contextID = contextID
	rt.resolver = NewResolver(client.Runtime, frameID)

	return rt
}

func (rt *Runtime) SetLoader(loader ValueLoader) *Runtime {
	rt.resolver.SetLoader(loader)

	return rt
}

func (rt *Runtime) ContextID() cdpruntime.ExecutionContextID {
	return rt.contextID
}

func (rt *Runtime) Eval(ctx context.Context, fn *Function) error {
	_, err := rt.evalInternal(ctx, fn.returnNothing())

	return err
}

func (rt *Runtime) EvalRef(ctx context.Context, fn *Function) (cdpruntime.RemoteObject, error) {
	out, err := rt.evalInternal(ctx, fn.returnRef())

	if err != nil {
		return cdpruntime.RemoteObject{}, err
	}

	return out, nil
}

func (rt *Runtime) EvalValue(ctx context.Context, fn *Function) (runtime.Value, error) {
	out, err := rt.evalInternal(ctx, fn.returnValue())

	if err != nil {
		return runtime.None, err
	}

	return rt.resolver.ToValue(ctx, out)
}

func (rt *Runtime) EvalElement(ctx context.Context, fn *Function) (runtime.Value, error) {
	ref, err := rt.EvalRef(ctx, fn)

	if err != nil {
		return nil, err
	}

	if ref.ObjectID == nil {
		return runtime.None, nil
	}

	return rt.resolver.ToElement(ctx, ref)
}

func (rt *Runtime) EvalElements(ctx context.Context, fn *Function) (*runtime.Array, error) {
	ref, err := rt.EvalRef(ctx, fn)

	if err != nil {
		return nil, err
	}

	val, err := rt.resolver.ToValue(ctx, ref)

	if err != nil {
		return nil, err
	}

	arr, ok := val.(*runtime.Array)

	if ok {
		return arr, nil
	}

	return runtime.NewArrayWith(val), nil
}

func (rt *Runtime) Compile(ctx context.Context, fn *Function) (*CompiledFunction, error) {
	log := rt.logger.With().
		Str("expression", fn.String()).
		Array("arguments", fn.args).
		Logger()

	arg := fn.compile(rt.contextID)

	log.Trace().Str("script", arg.Expression).Msg("compiling expression...")

	repl, err := rt.client.Runtime.CompileScript(ctx, arg)

	if err != nil {
		log.Trace().Err(err).Msg("failed compiling expression")

		return nil, err
	}

	if err := parseRuntimeException(repl.ExceptionDetails); err != nil {
		log.Trace().Err(err).Msg("compilation has failed with runtime exception")

		return nil, err
	}

	if repl.ScriptID == nil {
		log.Trace().Err(runtime.ErrUnexpected).Msg("compilation did not return script id")

		return nil, runtime.ErrUnexpected
	}

	id := *repl.ScriptID

	log.Trace().
		Str("script_id", string(id)).
		Msg("succeeded compiling expression")

	return CF(id, fn), nil
}

func (rt *Runtime) Call(ctx context.Context, fn *CompiledFunction) error {
	_, err := rt.callInternal(ctx, fn.returnNothing())

	return err
}

func (rt *Runtime) CallRef(ctx context.Context, fn *CompiledFunction) (cdpruntime.RemoteObject, error) {
	out, err := rt.callInternal(ctx, fn.returnRef())

	if err != nil {
		return cdpruntime.RemoteObject{}, err
	}

	return out, nil
}

func (rt *Runtime) CallValue(ctx context.Context, fn *CompiledFunction) (runtime.Value, error) {
	out, err := rt.callInternal(ctx, fn.returnValue())

	if err != nil {
		return runtime.None, err
	}

	return rt.resolver.ToValue(ctx, out)
}

func (rt *Runtime) CallElement(ctx context.Context, fn *CompiledFunction) (drivers.HTMLElement, error) {
	ref, err := rt.CallRef(ctx, fn)

	if err != nil {
		return nil, err
	}

	return rt.resolver.ToElement(ctx, ref)
}

func (rt *Runtime) CallElements(ctx context.Context, fn *CompiledFunction) (runtime.List, error) {
	ref, err := rt.CallRef(ctx, fn)

	if err != nil {
		return nil, err
	}

	val, err := rt.resolver.ToValue(ctx, ref)

	if err != nil {
		return nil, err
	}

	arr, ok := val.(runtime.List)

	if ok {
		return arr, nil
	}

	return runtime.NewArrayWith(val), nil
}

func (rt *Runtime) evalInternal(ctx context.Context, fn *Function) (cdpruntime.RemoteObject, error) {
	log := rt.logger.With().
		Str("expression", fn.String()).
		Str("returns", fn.returnType.String()).
		Bool("is_async", fn.async).
		Str("owner", string(fn.ownerID)).
		Array("arguments", fn.args).
		Logger()

	log.Trace().Msg("executing expression...")

	repl, err := rt.client.Runtime.CallFunctionOn(ctx, fn.eval(rt.contextID))

	if err != nil {
		log.Trace().Err(err).Msg("failed executing expression")

		return cdpruntime.RemoteObject{}, runtime.Error(err, "runtime evalInternal")
	}

	if err := parseRuntimeException(repl.ExceptionDetails); err != nil {
		log.Trace().Err(err).Msg("expression has failed with runtime exception")

		return cdpruntime.RemoteObject{}, err
	}

	var className string

	if repl.Result.ClassName != nil {
		className = *repl.Result.ClassName
	}

	var subtype string

	if repl.Result.Subtype != nil {
		subtype = *repl.Result.Subtype
	}

	log.Trace().
		Str("returned_type", repl.Result.Type).
		Str("returned_sub_type", subtype).
		Str("returned_class_name", className).
		Str("returned_value", string(repl.Result.Value)).
		Msg("succeeded executing expression")

	return repl.Result, nil
}

func (rt *Runtime) callInternal(ctx context.Context, fn *CompiledFunction) (cdpruntime.RemoteObject, error) {
	log := rt.logger.With().
		Str("script_id", string(fn.id)).
		Str("returns", fn.src.returnType.String()).
		Bool("is_async", fn.src.async).
		Array("arguments", fn.src.args).
		Logger()

	log.Trace().Msg("executing compiled script...")

	repl, err := rt.client.Runtime.RunScript(ctx, fn.call(rt.contextID))

	if err != nil {
		log.Trace().Err(err).Msg("failed executing compiled script")

		return cdpruntime.RemoteObject{}, runtime.Error(err, "runtime evalInternal")
	}

	if err := parseRuntimeException(repl.ExceptionDetails); err != nil {
		log.Trace().Err(err).Msg("compiled script has failed with runtime exception")

		return cdpruntime.RemoteObject{}, err
	}

	var className string

	if repl.Result.ClassName != nil {
		className = *repl.Result.ClassName
	}

	var subtype string

	if repl.Result.Subtype != nil {
		subtype = *repl.Result.Subtype
	}

	log.Trace().
		Str("returned_type", repl.Result.Type).
		Str("returned_sub_type", subtype).
		Str("returned_class_name", className).
		Str("returned_value", string(repl.Result.Value)).
		Msg("succeeded executing compiled script")

	return repl.Result, nil
}
