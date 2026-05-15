package eval

import (
	"context"
	"errors"
	"strconv"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/encoding/json"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type (
	ValueLoader interface {
		Load(
			ctx context.Context,
			frameID page.FrameID,
			remoteType RemoteObjectType,
			remoteClass RemoteClassName,
			id cdpruntime.RemoteObjectID,
		) (runtime.Value, error)
	}

	ValueLoaderFn func(
		ctx context.Context,
		frameID page.FrameID,
		remoteType RemoteObjectType,
		remoteClass RemoteClassName,
		id cdpruntime.RemoteObjectID,
	) (runtime.Value, error)

	Resolver struct {
		runtime cdp.Runtime
		loader  ValueLoader
		frameID page.FrameID
	}
)

func (f ValueLoaderFn) Load(
	ctx context.Context,
	frameID page.FrameID,
	remoteType RemoteObjectType,
	remoteClass RemoteClassName,
	id cdpruntime.RemoteObjectID,
) (runtime.Value, error) {
	return f(ctx, frameID, remoteType, remoteClass, id)
}

func NewResolver(runtime cdp.Runtime, frameID page.FrameID) *Resolver {
	return &Resolver{
		runtime: runtime,
		frameID: frameID,
		loader:  nil,
	}
}

func (r *Resolver) SetLoader(loader ValueLoader) *Resolver {
	r.loader = loader

	return r
}

func (r *Resolver) ToValue(ctx context.Context, ref cdpruntime.RemoteObject) (runtime.Value, error) {
	return r.toValue(ctx, ref, make(map[cdpruntime.RemoteObjectID]struct{}))
}

func (r *Resolver) ToList(ctx context.Context, ref cdpruntime.RemoteObject) (*runtime.Array, error) {
	val, err := r.toValue(ctx, ref, make(map[cdpruntime.RemoteObjectID]struct{}))
	if err != nil {
		return runtime.EmptyArray(), err
	}

	arr, ok := val.(*runtime.Array)
	if ok {
		return arr, nil
	}

	return runtime.NewArrayWith(val), nil
}

func (r *Resolver) toValue(
	ctx context.Context,
	ref cdpruntime.RemoteObject,
	seen map[cdpruntime.RemoteObjectID]struct{},
) (runtime.Value, error) {
	// It's not an actual ref but rather a plain value
	if ref.ObjectID == nil {
		if ref.Value != nil {
			return json.Default.Decode(ref.Value)
		}

		return runtime.None, nil
	}

	subtype := ToRemoteObjectType(ref)

	switch subtype {
	case NullObjectType, UndefinedObjectType:
		return runtime.None, nil
	case ArrayObjectType:
		return r.toArray(ctx, *ref.ObjectID, seen)
	case NodeObjectType:
		// is it even possible?
		if ref.ObjectID == nil {
			return json.Default.Decode(ref.Value)
		}

		return r.loadValue(ctx, NodeObjectType, ToRemoteClassName(ref), *ref.ObjectID)
	default:
		switch ToRemoteType(ref) {
		case StringType:
			str, err := strconv.Unquote(string(ref.Value))

			if err != nil {
				return runtime.None, err
			}

			return runtime.NewString(str), nil
		case ObjectType:
			if subtype == NullObjectType {
				return runtime.None, nil
			}

			if subtype == UnknownObjectType && r.isPlainObject(ref) {
				return r.toObject(ctx, *ref.ObjectID, seen)
			}

			return runtime.None, nil
		default:
			return json.Default.Decode(ref.Value)
		}
	}
}

func (r *Resolver) toArray(
	ctx context.Context,
	id cdpruntime.RemoteObjectID,
	seen map[cdpruntime.RemoteObjectID]struct{},
) (runtime.Value, error) {
	if _, ok := seen[id]; ok {
		return runtime.None, nil
	}

	seen[id] = struct{}{}

	props, err := r.getOwnProperties(ctx, id)
	if err != nil {
		return runtime.None, err
	}

	result := runtime.NewArray(len(props.Result))

	for _, descr := range props.Result {
		if !descr.Enumerable {
			continue
		}

		if descr.Value == nil {
			continue
		}

		el, err := r.toValue(ctx, *descr.Value, seen)
		if err != nil {
			return runtime.None, err
		}

		if err := result.Append(ctx, el); err != nil {
			return runtime.None, err
		}
	}

	return result, nil
}

func (r *Resolver) toObject(
	ctx context.Context,
	id cdpruntime.RemoteObjectID,
	seen map[cdpruntime.RemoteObjectID]struct{},
) (runtime.Value, error) {
	if _, ok := seen[id]; ok {
		return runtime.None, nil
	}

	seen[id] = struct{}{}

	props, err := r.getOwnProperties(ctx, id)
	if err != nil {
		return runtime.None, err
	}

	result := runtime.NewObject()

	for _, descr := range props.Result {
		if !descr.Enumerable {
			continue
		}

		if descr.Value == nil {
			continue
		}

		val, err := r.toValue(ctx, *descr.Value, seen)
		if err != nil {
			return runtime.None, err
		}

		if err := result.Set(ctx, runtime.NewString(descr.Name), val); err != nil {
			return runtime.None, err
		}
	}

	return result, nil
}

func (r *Resolver) getOwnProperties(
	ctx context.Context,
	id cdpruntime.RemoteObjectID,
) (*cdpruntime.GetPropertiesReply, error) {
	props, err := r.runtime.GetProperties(ctx, cdpruntime.NewGetPropertiesArgs(id).SetOwnProperties(true))
	if err != nil {
		return nil, err
	}

	if props.ExceptionDetails != nil {
		exception := *props.ExceptionDetails

		return nil, errors.New(exception.Text)
	}

	return props, nil
}

func (r *Resolver) isPlainObject(ref cdpruntime.RemoteObject) bool {
	if ToRemoteType(ref) != ObjectType {
		return false
	}

	if ToRemoteObjectType(ref) != UnknownObjectType {
		return false
	}

	if ref.ClassName == nil {
		return false
	}

	return *ref.ClassName == "Object"
}

func (r *Resolver) ToElement(ctx context.Context, ref cdpruntime.RemoteObject) (drivers.HTMLElement, error) {
	if ref.ObjectID == nil {
		return nil, runtime.Error(runtime.ErrInvalidArgument, "ref id")
	}

	val, err := r.loadValue(ctx, ToRemoteObjectType(ref), ToRemoteClassName(ref), *ref.ObjectID)

	if err != nil {
		return nil, err
	}

	return drivers.ToElement(val)
}

func (r *Resolver) ToProperty(
	ctx context.Context,
	id cdpruntime.RemoteObjectID,
	propName string,
) (runtime.Value, error) {
	res, err := r.runtime.GetProperties(
		ctx,
		cdpruntime.NewGetPropertiesArgs(id),
	)

	if err != nil {
		return runtime.None, err
	}

	if err := parseRuntimeException(res.ExceptionDetails); err != nil {
		return runtime.None, err
	}

	for _, prop := range res.Result {
		if prop.Name == propName {
			if prop.Value != nil {
				return r.ToValue(ctx, *prop.Value)
			}

			return runtime.None, nil
		}
	}

	return runtime.None, nil
}

func (r *Resolver) ToProperties(
	ctx context.Context,
	id cdpruntime.RemoteObjectID,
) (*runtime.Array, error) {
	res, err := r.runtime.GetProperties(
		ctx,
		cdpruntime.NewGetPropertiesArgs(id),
	)

	if err != nil {
		return runtime.EmptyArray(), err
	}

	if err := parseRuntimeException(res.ExceptionDetails); err != nil {
		return runtime.EmptyArray(), err
	}

	arr := runtime.NewArray(len(res.Result))

	for _, prop := range res.Result {
		if prop.Value != nil {
			val, err := r.ToValue(ctx, *prop.Value)

			if err != nil {
				return runtime.EmptyArray(), err
			}

			_ = arr.Append(ctx, val)
		}
	}

	return arr, nil
}

func (r *Resolver) loadValue(ctx context.Context, remoteType RemoteObjectType, remoteClass RemoteClassName, id cdpruntime.RemoteObjectID) (runtime.Value, error) {
	if r.loader == nil {
		return runtime.None, runtime.Error(runtime.ErrNotImplemented, "ValueLoader")
	}

	return r.loader.Load(ctx, r.frameID, remoteType, remoteClass, id)
}
