package core

import (
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeOpenOptionsRejectsUnknownFields(t *testing.T) {
	_, err := DecodeOpenOptions(t.Context(), runtime.NewObjectWith(map[string]runtime.Value{
		"memory": runtime.True,
		"extra":  runtime.True,
	}))
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}
