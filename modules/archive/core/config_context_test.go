package core

import (
	"context"
	"errors"
	"testing"
)

func TestConfigContextRoundTrip(t *testing.T) {
	t.Parallel()

	config := Config{
		MaxEntrySize:     123,
		MaxZIPBufferSize: 456,
	}
	ctx := WithConfig(nil, config)
	config.MaxEntrySize = 789

	got, err := ConfigFrom(ctx)
	if err != nil {
		t.Fatalf("unexpected config lookup error: %v", err)
	}
	if got != (Config{MaxEntrySize: 123, MaxZIPBufferSize: 456}) {
		t.Fatalf("unexpected config %#v", got)
	}
}

func TestConfigContextPreservesZeroConfig(t *testing.T) {
	t.Parallel()

	got, err := ConfigFrom(WithConfig(context.Background(), Config{}))
	if err != nil {
		t.Fatalf("unexpected config lookup error: %v", err)
	}
	if got != (Config{}) {
		t.Fatalf("expected zero config, got %#v", got)
	}
}

func TestConfigContextLatestValueWins(t *testing.T) {
	t.Parallel()

	parent := WithConfig(context.Background(), Config{
		MaxEntrySize:     1,
		MaxZIPBufferSize: 2,
	})
	child := WithConfig(parent, Config{
		MaxEntrySize:     3,
		MaxZIPBufferSize: 4,
	})

	got, err := ConfigFrom(child)
	if err != nil {
		t.Fatalf("unexpected config lookup error: %v", err)
	}
	if got != (Config{MaxEntrySize: 3, MaxZIPBufferSize: 4}) {
		t.Fatalf("unexpected config %#v", got)
	}
}

func TestConfigFromReturnsNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		ctx  context.Context
		name string
	}{
		{name: "nil context"},
		{name: "missing value", ctx: context.Background()},
		{
			name: "wrong value type",
			ctx:  context.WithValue(context.Background(), configContextKey{}, DefaultConfig().MaxEntrySize),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config, err := ConfigFrom(tt.ctx)
			if config != (Config{}) {
				t.Fatalf("expected zero config, got %#v", config)
			}
			if !errors.Is(err, ErrConfigNotFound) {
				t.Fatalf("expected ErrConfigNotFound, got %v", err)
			}
		})
	}
}
