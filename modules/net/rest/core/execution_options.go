package core

import (
	"context"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type ExecutionOptions struct {
	ResponseMode     ResponseMode
	RequestEncoding  Encoding
	ResponseEncoding Encoding
	ErrorMode        ErrorMode
	Timeout          time.Duration
}

func DefaultExecutionOptions(cfg Config) ExecutionOptions {
	return ExecutionOptions{
		Timeout:          time.Duration(cfg.Timeout),
		ResponseMode:     cfg.ResponseMode,
		RequestEncoding:  cfg.RequestEncoding,
		ResponseEncoding: cfg.ResponseEncoding,
		ErrorMode:        cfg.ErrorMode,
	}
}

func DecodeExecutionOptions(ctx context.Context, cfg Config, value runtime.Value) (ExecutionOptions, error) {
	opts := DefaultExecutionOptions(cfg)

	if runtime.TypeNone.Is(value) {
		return opts, nil
	}

	obj, err := requireMap(ctx, value, "HTTP query OPTIONS")
	if err != nil {
		return opts, err
	}

	if timeout, found, err := lookupDuration(ctx, obj, "timeout", "HTTP query OPTIONS"); err != nil {
		return opts, err
	} else if found {
		opts.Timeout = timeout
	}

	if response, found, err := lookupString(ctx, obj, "response", "HTTP query OPTIONS"); err != nil {
		return opts, err
	} else if found {
		opts.ResponseMode, err = parseResponseMode(response)
		if err != nil {
			return opts, OperationError("OPTIONS", err)
		}
	}

	if encoding, found, err := lookupString(ctx, obj, "requestEncoding", "HTTP query OPTIONS"); err != nil {
		return opts, err
	} else if found {
		opts.RequestEncoding, err = parseEncoding(encoding)
		if err != nil {
			return opts, OperationError("OPTIONS", err)
		}
	}

	if encoding, found, err := lookupString(ctx, obj, "responseEncoding", "HTTP query OPTIONS"); err != nil {
		return opts, err
	} else if found {
		opts.ResponseEncoding, err = parseEncoding(encoding)
		if err != nil {
			return opts, OperationError("OPTIONS", err)
		}
	}

	if errorMode, found, err := lookupString(ctx, obj, "errorMode", "HTTP query OPTIONS"); err != nil {
		return opts, err
	} else if found {
		opts.ErrorMode, err = parseErrorMode(errorMode)
		if err != nil {
			return opts, OperationError("OPTIONS", err)
		}
	}

	return opts, nil
}
