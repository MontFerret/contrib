package lib

import (
	"context"
	"fmt"
	"time"

	"github.com/MontFerret/contrib/modules/web/sitemap/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func parseOptions(ctx context.Context, args []runtime.Value) (core.Options, error) {
	opts := core.DefaultOptions()

	if len(args) < 2 {
		return opts, nil
	}

	m, err := runtime.CastArgAt[runtime.Map](args, 1)
	if err != nil {
		return opts, err
	}

	if err := m.ForEach(ctx, func(ctx context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		name, err := optionName(key)
		if err != nil {
			return false, err
		}

		switch name {
		case "recursive":
			raw, ok := value.(runtime.Boolean)
			if !ok {
				return false, fmt.Errorf(`sitemap option "recursive" must be a boolean`)
			}

			opts.Recursive = bool(raw)
		case "dedupe":
			raw, ok := value.(runtime.Boolean)
			if !ok {
				return false, fmt.Errorf(`sitemap option "dedupe" must be a boolean`)
			}

			opts.Dedupe = bool(raw)
		case "maxDepth":
			raw, ok := value.(runtime.Int)
			if !ok {
				return false, fmt.Errorf(`sitemap option "maxDepth" must be an integer`)
			}

			if raw < 0 {
				return false, fmt.Errorf(`sitemap option "maxDepth" must be >= 0`)
			}

			opts.MaxDepth = int(raw)
		case "timeout":
			raw, ok := value.(runtime.Int)
			if !ok {
				return false, fmt.Errorf(`sitemap option "timeout" must be an integer`)
			}

			if raw < 0 {
				return false, fmt.Errorf(`sitemap option "timeout" must be >= 0`)
			}

			opts.Timeout = time.Duration(raw) * time.Millisecond
		case "headers":
			headers, err := runtime.CastMap(value)
			if err != nil {
				return false, fmt.Errorf(`sitemap option "headers" must be an object`)
			}

			parsed, err := parseHeaders(ctx, headers)
			if err != nil {
				return false, err
			}

			opts.Headers = parsed
		default:
			return false, fmt.Errorf("unknown sitemap option %q", name)
		}

		return true, nil
	}); err != nil {
		return opts, err
	}

	return opts, nil
}

func optionName(key runtime.Value) (string, error) {
	str, ok := key.(runtime.String)
	if !ok {
		return "", fmt.Errorf("sitemap option keys must be strings")
	}

	return str.String(), nil
}

func parseHeaders(ctx context.Context, headers runtime.Map) (map[string]string, error) {
	out := make(map[string]string)

	if err := headers.ForEach(ctx, func(_ context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		name, ok := key.(runtime.String)
		if !ok {
			return false, fmt.Errorf("sitemap headers keys must be strings")
		}

		text, ok := value.(runtime.String)
		if !ok {
			return false, fmt.Errorf("sitemap headers values must be strings")
		}

		out[name.String()] = text.String()

		return true, nil
	}); err != nil {
		return nil, err
	}

	return out, nil
}
