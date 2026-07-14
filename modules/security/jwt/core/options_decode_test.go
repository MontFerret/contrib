package core

import (
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestOptionDecodersRejectUnknownFields(t *testing.T) {
	verify := runtime.NewObjectWith(map[string]runtime.Value{
		"algorithms": runtime.NewArrayWith(runtime.NewString("HS256")),
		"extra":      runtime.True,
	})
	if _, err := DecodeVerifyOptions(t.Context(), verify); err == nil {
		t.Fatalf("expected verify unknown field error, got %v", err)
	}

	sign := runtime.NewObjectWith(map[string]runtime.Value{
		"algorithm": runtime.NewString("HS256"),
		"extra":     runtime.True,
	})
	if _, err := DecodeSignOptions(t.Context(), sign); err == nil {
		t.Fatalf("expected sign unknown field error, got %v", err)
	}
}
