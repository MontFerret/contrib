package core

import (
	"context"
	"strings"
	"sync"
)

// Registry resolves normalized provider names to factories.
type Registry struct {
	factories map[string]ProviderFactory
	mu        sync.RWMutex
}

// NewRegistry creates an empty provider registry.
func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]ProviderFactory)}
}

// Register installs a provider factory under its normalized name.
func (r *Registry) Register(factory ProviderFactory) error {
	name, err := validateFactory(factory)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return NewError(ErrInvalidOptions, "provider is already registered")
	}

	r.factories[name] = factory

	return nil
}

// Replace installs a provider factory, replacing any factory under the same name.
func (r *Registry) Replace(factory ProviderFactory) error {
	name, err := validateFactory(factory)
	if err != nil {
		return err
	}

	r.mu.Lock()
	r.factories[name] = factory
	r.mu.Unlock()

	return nil
}

// NewModel creates a stateless model through a registered provider factory.
func (r *Registry) NewModel(ctx context.Context, provider string, options ModelOptions) (Model, error) {
	name := strings.ToLower(strings.TrimSpace(provider))

	r.mu.RLock()
	factory, found := r.factories[name]
	r.mu.RUnlock()

	if !found {
		return nil, NewError(ErrUnsupportedProvider, "unsupported provider")
	}

	return factory.NewModel(ctx, options)
}
