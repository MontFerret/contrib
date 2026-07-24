package core

import (
	"context"
	"errors"
	"testing"
)

func TestRegistryContextRoundTrip(t *testing.T) {
	registry := NewRegistry()
	ctx := WithRegistry(nil, registry)

	actual, err := RegistryFrom(ctx)
	if err != nil {
		t.Fatalf("unexpected registry lookup error: %v", err)
	}
	if actual != registry {
		t.Fatalf("expected registry %p, got %p", registry, actual)
	}
}

func TestRegistryFromReturnsNotFound(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{name: "nil context"},
		{name: "missing value", ctx: context.Background()},
		{
			name: "wrong value type",
			ctx:  context.WithValue(context.Background(), registryCtxKey, "not a registry"),
		},
		{name: "nil registry", ctx: WithRegistry(context.Background(), nil)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registry, err := RegistryFrom(test.ctx)
			if registry != nil {
				t.Fatalf("expected no registry, got %p", registry)
			}
			if !errors.Is(err, ErrRegistryNotFound) {
				t.Fatalf("expected ErrRegistryNotFound, got %v", err)
			}
		})
	}
}
