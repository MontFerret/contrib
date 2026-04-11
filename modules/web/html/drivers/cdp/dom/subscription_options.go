package dom

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const (
	domEventOptionDelegate       = "delegate"
	domEventOptionListener       = "listener"
	domEventOptionMaxDepth       = "maxDepth"
	domEventOptionProps          = "props"
	domEventOptionTargetSelector = "targetSelector"

	defaultDOMEventMaxDepth = 4
	minDOMEventMaxDepth     = 1
	maxDOMEventMaxDepth     = 8
)

type domEventOptions struct {
	Listener          runtime.Map
	Props             *runtime.Array
	Delegate          runtime.String
	TargetSelector    runtime.String
	MaxDepth          runtime.Int
	HasDelegate       bool
	HasTargetSelector bool
	HasProps          bool
}

func defaultDOMEventOptions() domEventOptions {
	return domEventOptions{
		MaxDepth: runtime.NewInt(defaultDOMEventMaxDepth),
	}
}

func parseDOMEventOptions(ctx context.Context, value runtime.Map) (domEventOptions, error) {
	options := defaultDOMEventOptions()

	if value == nil {
		return options, nil
	}

	if err := validateDOMEventOptionKeys(ctx, value); err != nil {
		return options, err
	}

	listenerValue, err := value.Get(ctx, runtime.NewString(domEventOptionListener))

	if err != nil {
		return options, err
	}

	if listenerValue != runtime.None {
		listener, err := runtime.CastMap(listenerValue)

		if err != nil {
			return options, runtime.Errorf(runtime.ErrInvalidArgument, "%s: %s", domEventOptionListener, err)
		}

		options.Listener = listener
	}

	delegateValue, err := value.Get(ctx, runtime.NewString(domEventOptionDelegate))

	if err != nil {
		return options, err
	}

	if delegateValue != runtime.None {
		delegate, err := runtime.CastString(delegateValue)

		if err != nil {
			return options, runtime.Errorf(runtime.ErrInvalidArgument, "%s: %s", domEventOptionDelegate, err)
		}

		options.Delegate = delegate
		options.HasDelegate = true
	}

	targetSelectorValue, err := value.Get(ctx, runtime.NewString(domEventOptionTargetSelector))

	if err != nil {
		return options, err
	}

	if targetSelectorValue != runtime.None {
		targetSelector, err := runtime.CastString(targetSelectorValue)

		if err != nil {
			return options, runtime.Errorf(runtime.ErrInvalidArgument, "%s: %s", domEventOptionTargetSelector, err)
		}

		options.TargetSelector = targetSelector
		options.HasTargetSelector = true
	}

	if options.HasDelegate && options.HasTargetSelector {
		return options, runtime.Error(runtime.ErrInvalidArgument, "delegate and targetSelector cannot be used together")
	}

	propsValue, err := value.Get(ctx, runtime.NewString(domEventOptionProps))

	if err != nil {
		return options, err
	}

	if propsValue != runtime.None {
		props, err := runtime.CastList(propsValue)

		if err != nil {
			return options, runtime.Errorf(runtime.ErrInvalidArgument, "%s: %s", domEventOptionProps, err)
		}

		options.Props, err = normalizeDOMEventProps(ctx, props)

		if err != nil {
			return options, err
		}

		options.HasProps = true
	}

	maxDepthValue, err := value.Get(ctx, runtime.NewString(domEventOptionMaxDepth))

	if err != nil {
		return options, err
	}

	if maxDepthValue != runtime.None {
		maxDepth, err := runtime.CastInt(maxDepthValue)

		if err != nil {
			return options, runtime.Errorf(runtime.ErrInvalidArgument, "%s: %s", domEventOptionMaxDepth, err)
		}

		if maxDepth < runtime.NewInt(minDOMEventMaxDepth) || maxDepth > runtime.NewInt(maxDOMEventMaxDepth) {
			return options, runtime.Errorf(
				runtime.ErrRange,
				"%s must be between %d and %d",
				domEventOptionMaxDepth,
				minDOMEventMaxDepth,
				maxDOMEventMaxDepth,
			)
		}

		options.MaxDepth = maxDepth
	}

	return options, nil
}

func validateDOMEventOptionKeys(ctx context.Context, value runtime.Map) error {
	keys, err := value.Keys(ctx)

	if err != nil {
		return err
	}

	length, err := keys.Length(ctx)

	if err != nil {
		return err
	}

	for idx := runtime.NewInt(0); idx < length; idx++ {
		keyValue, err := keys.At(ctx, idx)

		if err != nil {
			return err
		}

		key, err := runtime.CastString(keyValue)

		if err != nil {
			return runtime.Errorf(runtime.ErrInvalidArgument, "options key: %s", err)
		}

		switch key {
		case domEventOptionDelegate, domEventOptionListener, domEventOptionMaxDepth, domEventOptionProps, domEventOptionTargetSelector:
			continue
		default:
			return runtime.Errorf(runtime.ErrInvalidArgument, "unknown DOM event option: %s", key)
		}
	}

	return nil
}

func normalizeDOMEventProps(ctx context.Context, value runtime.List) (*runtime.Array, error) {
	length, err := value.Length(ctx)

	if err != nil {
		return nil, err
	}

	result := runtime.NewArray(int(length))
	seen := make(map[string]struct{}, int(length))

	for idx := runtime.NewInt(0); idx < length; idx++ {
		item, err := value.At(ctx, idx)

		if err != nil {
			return nil, err
		}

		prop, err := runtime.CastString(item)

		if err != nil {
			return nil, runtime.Errorf(runtime.ErrInvalidArgument, "%s[%d]: %s", domEventOptionProps, idx, err)
		}

		name := prop.String()

		if _, ok := seen[name]; ok {
			continue
		}

		seen[name] = struct{}{}

		if err := result.Append(ctx, prop); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func buildDOMEventTemplateOptions(options domEventOptions) runtime.Map {
	config := runtime.NewObjectWith(map[string]runtime.Value{
		domEventOptionListener:       runtime.None,
		domEventOptionDelegate:       runtime.None,
		domEventOptionTargetSelector: runtime.None,
		domEventOptionProps:          runtime.None,
		domEventOptionMaxDepth:       options.MaxDepth,
	})

	if options.Listener != nil {
		_ = config.Set(context.Background(), runtime.NewString(domEventOptionListener), options.Listener)
	}

	if options.HasDelegate {
		_ = config.Set(context.Background(), runtime.NewString(domEventOptionDelegate), options.Delegate)
	}

	if options.HasTargetSelector {
		_ = config.Set(context.Background(), runtime.NewString(domEventOptionTargetSelector), options.TargetSelector)
	}

	if options.HasProps {
		_ = config.Set(context.Background(), runtime.NewString(domEventOptionProps), options.Props)
	}

	return config
}
