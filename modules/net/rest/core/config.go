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

const clientConfigOwner = "NET::REST::CLIENT config"

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

	obj, err := requireMap(ctx, value, clientConfigOwner)
	if err != nil {
		return cfg, fmt.Errorf("%s or a string", err.Error())
	}

	if baseURL, found, err := lookupString(ctx, obj, "baseUrl", clientConfigOwner); err != nil {
		return cfg, err
	} else if found {
		cfg.BaseURL = baseURL
	}
	if cfg.BaseURL == "" {
		return cfg, fmt.Errorf("%s.baseUrl is required", clientConfigOwner)
	}

	if headers, found, err := lookupValue(ctx, obj, "headers"); err != nil {
		return cfg, err
	} else if found {
		cfg.Headers, err = decodeHeaders(ctx, headers, clientConfigOwner+".headers")
		if err != nil {
			return cfg, err
		}
	}

	if encoding, found, err := lookupString(ctx, obj, "encoding", clientConfigOwner); err != nil {
		return cfg, err
	} else if found {
		enc, err := parseEncoding(encoding)
		if err != nil {
			return cfg, fmt.Errorf("%s.encoding: %w", clientConfigOwner, err)
		}

		cfg.RequestEncoding = enc
		cfg.ResponseEncoding = enc
	}

	if encoding, found, err := lookupString(ctx, obj, "requestEncoding", clientConfigOwner); err != nil {
		return cfg, err
	} else if found {
		cfg.RequestEncoding, err = parseEncoding(encoding)
		if err != nil {
			return cfg, fmt.Errorf("%s.requestEncoding: %w", clientConfigOwner, err)
		}
	}

	if encoding, found, err := lookupString(ctx, obj, "responseEncoding", clientConfigOwner); err != nil {
		return cfg, err
	} else if found {
		cfg.ResponseEncoding, err = parseEncoding(encoding)
		if err != nil {
			return cfg, fmt.Errorf("%s.responseEncoding: %w", clientConfigOwner, err)
		}
	}

	if timeout, found, err := lookupDuration(ctx, obj, "timeout", clientConfigOwner); err != nil {
		return cfg, err
	} else if found {
		cfg.Timeout = int64(timeout)
	}

	if response, found, err := lookupString(ctx, obj, "response", clientConfigOwner); err != nil {
		return cfg, err
	} else if found {
		cfg.ResponseMode, err = parseResponseMode(response)
		if err != nil {
			return cfg, fmt.Errorf("%s.response: %w", clientConfigOwner, err)
		}
	}

	if errorMode, found, err := lookupString(ctx, obj, "errorMode", clientConfigOwner); err != nil {
		return cfg, err
	} else if found {
		cfg.ErrorMode, err = parseErrorMode(errorMode)
		if err != nil {
			return cfg, fmt.Errorf("%s.errorMode: %w", clientConfigOwner, err)
		}
	}

	return cfg, nil
}
