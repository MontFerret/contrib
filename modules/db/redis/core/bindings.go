package core

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func compileQueryTemplate(
	ctx context.Context,
	input string,
	bindingsValue runtime.Value,
) (string, []any, error) {
	tokens, err := parseQueryTemplate(input)
	if err != nil {
		return "", nil, err
	}

	bindings, err := parseBindings(bindingsValue)
	if err != nil {
		return "", nil, err
	}

	command, err := resolveCommandToken(tokens[0])
	if err != nil {
		return "", nil, err
	}

	args := make([]any, 0, len(tokens)-1)
	for _, token := range tokens[1:] {
		resolved, err := resolveArgumentToken(ctx, token, bindings)
		if err != nil {
			return "", nil, err
		}

		args = append(args, resolved...)
	}

	return command, args, nil
}

func parseBindings(input runtime.Value) (runtime.Map, error) {
	if input == nil || input == runtime.None {
		return nil, nil
	}

	bindings, ok := input.(runtime.Map)
	if !ok {
		return nil, fmt.Errorf("query WITH bindings must be an object")
	}

	return bindings, nil
}

func resolveCommandToken(token queryTemplateToken) (string, error) {
	var command strings.Builder
	for _, part := range token.parts {
		if part.binding != "" {
			return "", fmt.Errorf("redis command must be static; binding $%s is not allowed in command position", part.binding)
		}

		command.WriteString(part.literal)
	}

	if command.Len() == 0 {
		return "", fmt.Errorf("redis query must contain a non-empty command")
	}

	return command.String(), nil
}

func resolveArgumentToken(
	ctx context.Context,
	token queryTemplateToken,
	bindings runtime.Map,
) ([]any, error) {
	if len(token.parts) == 1 && token.parts[0].binding != "" {
		part := token.parts[0]
		value, err := lookupBinding(ctx, bindings, part.binding)
		if err != nil {
			return nil, err
		}

		if part.spread {
			return resolveSpreadBinding(ctx, part.binding, value)
		}

		arg, err := runtimeValueToRedisArg(value)
		if err != nil {
			return nil, fmt.Errorf("redis query binding $%s: %w", part.binding, err)
		}

		return []any{arg}, nil
	}

	var composed strings.Builder
	for _, part := range token.parts {
		if part.binding == "" {
			composed.WriteString(part.literal)
			continue
		}

		value, err := lookupBinding(ctx, bindings, part.binding)
		if err != nil {
			return nil, err
		}

		text, err := runtimeValueToRedisText(value)
		if err != nil {
			return nil, fmt.Errorf("redis query binding $%s: %w", part.binding, err)
		}

		composed.WriteString(text)
	}

	return []any{composed.String()}, nil
}

func lookupBinding(ctx context.Context, bindings runtime.Map, name string) (runtime.Value, error) {
	if bindings == nil {
		return runtime.None, fmt.Errorf("missing Redis query binding %q", name)
	}

	value, found, err := bindings.Lookup(ctx, runtime.NewString(name))
	if err != nil {
		return runtime.None, err
	}
	if !found {
		return runtime.None, fmt.Errorf("missing Redis query binding %q", name)
	}

	return value, nil
}

func resolveSpreadBinding(ctx context.Context, name string, value runtime.Value) ([]any, error) {
	values, ok := value.(runtime.List)
	if !ok {
		return nil, fmt.Errorf("redis query spread binding $%s... must be an array or list", name)
	}

	args := make([]any, 0)
	index := 0

	if err := runtime.ForEach(ctx, values, func(_ context.Context, value, _ runtime.Value) (runtime.Boolean, error) {
		arg, err := runtimeValueToRedisArg(value)
		if err != nil {
			return runtime.False, fmt.Errorf("redis query spread binding $%s...[%d]: %w", name, index, err)
		}

		args = append(args, arg)
		index++

		return runtime.True, nil
	}); err != nil {
		return nil, err
	}

	return args, nil
}

func runtimeValueToRedisArg(value runtime.Value) (any, error) {
	if value == nil || value == runtime.None {
		return nil, fmt.Errorf("NONE is not a supported Redis command argument")
	}

	switch val := value.(type) {
	case runtime.Int:
		return int64(val), nil
	case runtime.Float:
		return float64(val), nil
	case runtime.String:
		return val.String(), nil
	case runtime.Boolean:
		return bool(val), nil
	case runtime.Binary:
		out := make([]byte, len(val))
		copy(out, val)

		return out, nil
	default:
		return nil, fmt.Errorf("unsupported Redis argument type %s", runtime.TypeName(runtime.TypeOf(value)))
	}
}

func runtimeValueToRedisText(value runtime.Value) (string, error) {
	arg, err := runtimeValueToRedisArg(value)
	if err != nil {
		return "", err
	}

	switch val := arg.(type) {
	case int64:
		return strconv.FormatInt(val, 10), nil
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64), nil
	case string:
		return val, nil
	case bool:
		if val {
			return "1", nil
		}

		return "0", nil
	case []byte:
		return string(val), nil
	default:
		return "", fmt.Errorf("unsupported Redis argument type %T", arg)
	}
}
