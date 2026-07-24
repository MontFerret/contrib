package core

import (
	"context"
	"errors"
)

// ErrRegistryNotFound indicates that no provider registry is available in the context.
var ErrRegistryNotFound = errors.New("ai/llm: provider registry not found in context")

type registryContextKey struct{}

var registryCtxKey = registryContextKey{}

// WithRegistry adds a provider registry to the context.
func WithRegistry(ctx context.Context, registry *Registry) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, registryCtxKey, registry)
}

// RegistryFrom returns the provider registry stored in the context.
func RegistryFrom(ctx context.Context) (*Registry, error) {
	if ctx == nil {
		return nil, ErrRegistryNotFound
	}

	registry, ok := ctx.Value(registryCtxKey).(*Registry)
	if !ok || registry == nil {
		return nil, ErrRegistryNotFound
	}

	return registry, nil
}
