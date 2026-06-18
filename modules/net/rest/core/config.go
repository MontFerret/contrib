package core

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type Config struct {
	BaseURL          string
	Headers          http.Header
	RequestEncoding  Encoding
	ResponseEncoding Encoding
	ResponseMode     ResponseMode
	ErrorMode        ErrorMode
	Timeout          int64
}

func DefaultConfig() Config {
	return Config{
		Headers:          make(http.Header),
		RequestEncoding:  EncodingJSON,
		ResponseEncoding: EncodingJSON,
		ResponseMode:     ResponseModeBody,
		ErrorMode:        ErrorModeRaise,
	}
}

func DecodeClientConfig(ctx context.Context, value runtime.Value) (Config, error) {
	cfg := DefaultConfig()

	if runtime.TypeOf(value) == runtime.TypeString {
		cfg.BaseURL = value.String()
		return cfg, nil
	}

	obj, err := requireMap(ctx, value, "HTTP::CLIENT config")
	if err != nil {
		return cfg, fmt.Errorf("%s or a string", err.Error())
	}

	if baseURL, found, err := lookupString(ctx, obj, "baseUrl", "HTTP::CLIENT config"); err != nil {
		return cfg, err
	} else if found {
		cfg.BaseURL = baseURL
	}
	if cfg.BaseURL == "" {
		return cfg, fmt.Errorf("HTTP::CLIENT config.baseUrl is required")
	}

	if headers, found, err := lookupValue(ctx, obj, "headers"); err != nil {
		return cfg, err
	} else if found {
		cfg.Headers, err = decodeHeaders(ctx, headers, "HTTP::CLIENT config.headers")
		if err != nil {
			return cfg, err
		}
	}

	if encoding, found, err := lookupString(ctx, obj, "encoding", "HTTP::CLIENT config"); err != nil {
		return cfg, err
	} else if found {
		enc, err := parseEncoding(encoding)
		if err != nil {
			return cfg, fmt.Errorf("HTTP::CLIENT config.encoding: %w", err)
		}

		cfg.RequestEncoding = enc
		cfg.ResponseEncoding = enc
	}

	if encoding, found, err := lookupString(ctx, obj, "requestEncoding", "HTTP::CLIENT config"); err != nil {
		return cfg, err
	} else if found {
		cfg.RequestEncoding, err = parseEncoding(encoding)
		if err != nil {
			return cfg, fmt.Errorf("HTTP::CLIENT config.requestEncoding: %w", err)
		}
	}

	if encoding, found, err := lookupString(ctx, obj, "responseEncoding", "HTTP::CLIENT config"); err != nil {
		return cfg, err
	} else if found {
		cfg.ResponseEncoding, err = parseEncoding(encoding)
		if err != nil {
			return cfg, fmt.Errorf("HTTP::CLIENT config.responseEncoding: %w", err)
		}
	}

	if timeout, found, err := lookupDuration(ctx, obj, "timeout", "HTTP::CLIENT config"); err != nil {
		return cfg, err
	} else if found {
		cfg.Timeout = int64(timeout)
	}

	if response, found, err := lookupString(ctx, obj, "response", "HTTP::CLIENT config"); err != nil {
		return cfg, err
	} else if found {
		cfg.ResponseMode, err = parseResponseMode(response)
		if err != nil {
			return cfg, fmt.Errorf("HTTP::CLIENT config.response: %w", err)
		}
	}

	if errorMode, found, err := lookupString(ctx, obj, "errorMode", "HTTP::CLIENT config"); err != nil {
		return cfg, err
	} else if found {
		cfg.ErrorMode, err = parseErrorMode(errorMode)
		if err != nil {
			return cfg, fmt.Errorf("HTTP::CLIENT config.errorMode: %w", err)
		}
	}

	return cfg, nil
}
