package network

import (
	"context"
	"time"

	cdpnetwork "github.com/mafredri/cdp/protocol/network"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const (
	defaultNetworkBodyLimit = 1048576
	defaultNetworkIdleQuiet = 500 * time.Millisecond
)

type (
	networkEventOptions struct {
		bodyLimit   int
		captureBody bool
	}

	networkIdleOptions struct {
		types       map[string]struct{}
		typeList    []string
		quiet       time.Duration
		maxInflight int
	}
)

func parseNetworkEventOptions(ctx context.Context, eventName string, options runtime.Map) (networkEventOptions, error) {
	result := networkEventOptions{
		bodyLimit: defaultNetworkBodyLimit,
	}

	if options == nil {
		return result, nil
	}

	err := options.ForEach(ctx, func(_ context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		switch key.String() {
		case "captureBody":
			if eventName != drivers.NetworkRequestFinishedEvent {
				return runtime.False, unknownOptionError(eventName, key.String())
			}

			captureBody, err := runtime.CastBoolean(value)
			if err != nil {
				return runtime.False, optionTypeError(eventName, key.String(), "boolean")
			}

			result.captureBody = bool(captureBody)
		case "bodyLimit":
			if eventName != drivers.NetworkRequestFinishedEvent {
				return runtime.False, unknownOptionError(eventName, key.String())
			}

			bodyLimit, err := parseIntegerOption(eventName, key.String(), value)
			if err != nil {
				return runtime.False, err
			}

			if bodyLimit < 0 {
				return runtime.False, runtime.Errorf(
					runtime.ErrInvalidArgument,
					"%s option %q must be greater than or equal to 0",
					eventName,
					key.String(),
				)
			}

			result.bodyLimit = bodyLimit
		default:
			return runtime.False, unknownOptionError(eventName, key.String())
		}

		return runtime.True, nil
	})

	return result, err
}

func parseNetworkIdleOptions(ctx context.Context, eventName string, options runtime.Map) (networkIdleOptions, error) {
	result := networkIdleOptions{
		quiet: defaultNetworkIdleQuiet,
	}

	if options == nil {
		return result, nil
	}

	err := options.ForEach(ctx, func(ctx context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		switch key.String() {
		case "quiet":
			quiet, err := parseDurationOption(eventName, key.String(), value)
			if err != nil {
				return runtime.False, err
			}

			if quiet < 0 {
				return runtime.False, runtime.Errorf(
					runtime.ErrInvalidArgument,
					"%s option %q must be greater than or equal to 0",
					eventName,
					key.String(),
				)
			}

			result.quiet = quiet
		case "maxInflight":
			maxInflight, err := parseIntegerOption(eventName, key.String(), value)
			if err != nil {
				return runtime.False, err
			}

			if maxInflight < 0 {
				return runtime.False, runtime.Errorf(
					runtime.ErrInvalidArgument,
					"%s option %q must be greater than or equal to 0",
					eventName,
					key.String(),
				)
			}

			result.maxInflight = maxInflight
		case "types":
			types, typeList, err := parseResourceTypesOption(ctx, eventName, key.String(), value)
			if err != nil {
				return runtime.False, err
			}

			result.types = types
			result.typeList = typeList
		default:
			return runtime.False, unknownOptionError(eventName, key.String())
		}

		return runtime.True, nil
	})

	return result, err
}

func parseIntegerOption(eventName, optionName string, value runtime.Value) (int, error) {
	switch typed := value.(type) {
	case runtime.Int:
		return int(typed), nil
	case runtime.Float:
		return int(typed), nil
	default:
		return 0, optionTypeError(eventName, optionName, "number")
	}
}

func parseDurationOption(eventName, optionName string, value runtime.Value) (time.Duration, error) {
	switch typed := value.(type) {
	case runtime.Int:
		return time.Duration(typed) * time.Millisecond, nil
	case runtime.Float:
		return time.Duration(float64(typed) * float64(time.Millisecond)), nil
	case runtime.String:
		duration, err := time.ParseDuration(typed.String())
		if err != nil {
			return 0, runtime.Errorf(
				runtime.ErrInvalidArgument,
				"%s option %q must be a duration string or number of milliseconds",
				eventName,
				optionName,
			)
		}

		return duration, nil
	default:
		return 0, optionTypeError(eventName, optionName, "duration")
	}
}

func parseResourceTypesOption(
	ctx context.Context,
	eventName string,
	optionName string,
	value runtime.Value,
) (map[string]struct{}, []string, error) {
	arr, err := runtime.CastArray(value)
	if err != nil {
		return nil, nil, optionTypeError(eventName, optionName, "array")
	}

	length, err := arr.Length(ctx)
	if err != nil {
		return nil, nil, err
	}

	if length == 0 {
		return nil, nil, nil
	}

	types := make(map[string]struct{}, int(length))
	typeList := make([]string, 0, int(length))

	for i := runtime.Int(0); i < length; i++ {
		item, err := arr.At(ctx, i)
		if err != nil {
			return nil, nil, err
		}

		itemValue, err := runtime.CastString(item)
		if err != nil {
			return nil, nil, optionTypeError(eventName, optionName, "array of strings")
		}

		resourceType := normalizeResourceTypeAlias(itemValue.String())
		if resourceType == "" || toResourceType(resourceType) == cdpnetwork.ResourceTypeNotSet {
			return nil, nil, runtime.Errorf(
				runtime.ErrInvalidArgument,
				"%s option %q contains unsupported resource type: %s",
				eventName,
				optionName,
				itemValue,
			)
		}

		if _, exists := types[resourceType]; exists {
			continue
		}

		types[resourceType] = struct{}{}
		typeList = append(typeList, resourceType)
	}

	return types, typeList, nil
}

func unknownOptionError(eventName, optionName string) error {
	return runtime.Errorf(runtime.ErrInvalidArgument, "unknown %s option: %s", eventName, optionName)
}

func optionTypeError(eventName, optionName, expected string) error {
	return runtime.Errorf(
		runtime.ErrInvalidArgument,
		"%s option %q must be %s",
		eventName,
		optionName,
		expected,
	)
}

func invalidNetworkEventNameError(eventName string) error {
	return runtime.Errorf(runtime.ErrInvalidOperation, "unknown event name: %s", eventName)
}
