package core

import (
	"context"
	"errors"
)

type configContextKey struct{}

// ErrConfigNotFound indicates that archive configuration is unavailable in
// the execution context.
var ErrConfigNotFound = errors.New("archive: config not found in context")

// WithConfig adds archive configuration to the execution context.
func WithConfig(ctx context.Context, config Config) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, configContextKey{}, config)
}

// ConfigFrom returns archive configuration from the execution context.
func ConfigFrom(ctx context.Context) (Config, error) {
	if ctx == nil {
		return Config{}, ErrConfigNotFound
	}

	config, ok := ctx.Value(configContextKey{}).(Config)
	if !ok {
		return Config{}, ErrConfigNotFound
	}

	return config, nil
}
