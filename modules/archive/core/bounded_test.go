package core

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func TestReadBoundedEnforcesObservedSize(t *testing.T) {
	t.Parallel()

	_, err := readBounded(context.Background(), bytes.NewBufferString("12345"), "entry", 4)
	var limitErr *LimitError

	if !errors.As(err, &limitErr) {
		t.Fatalf("expected LimitError, got %v", err)
	}

	if limitErr.Limit != 4 {
		t.Fatalf("unexpected limit error %#v", limitErr)
	}
}

func TestReadBoundedHonorsCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := readBounded(ctx, bytes.NewBufferString("data"), "entry", 10); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
