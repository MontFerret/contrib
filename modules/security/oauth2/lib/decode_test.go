package lib

import (
	"context"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeParametersSupportsScalarsRepeatedValuesAndNone(t *testing.T) {
	t.Parallel()

	parameters, err := decodeParameters(
		context.Background(),
		runtime.NewObjectWith(map[string]runtime.Value{
			"text": runtime.NewString("value"),
			"int":  runtime.NewInt(42),
			"bool": runtime.True,
			"many": runtime.NewArrayWith(
				runtime.NewString("first"),
				runtime.None,
				runtime.NewString("second"),
			),
			"omitted": runtime.None,
		}),
		"parameters",
	)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}

	if got := parameters["text"]; len(got) != 1 || got[0] != "value" {
		t.Fatalf("text = %v", got)
	}
	if got := parameters["int"]; len(got) != 1 || got[0] != "42" {
		t.Fatalf("int = %v", got)
	}
	if got := parameters["bool"]; len(got) != 1 || got[0] != "true" {
		t.Fatalf("bool = %v", got)
	}
	if got := parameters["many"]; len(got) != 2 || got[0] != "first" || got[1] != "second" {
		t.Fatalf("many = %v", got)
	}
	if _, found := parameters["omitted"]; found {
		t.Fatal("NONE parameter must be omitted")
	}
}

func TestDecodeParametersRejectsNestedValues(t *testing.T) {
	t.Parallel()

	_, err := decodeParameters(
		context.Background(),
		runtime.NewObjectWith(map[string]runtime.Value{
			"nested": runtime.NewObject(),
		}),
		"parameters",
	)
	if err == nil || !strings.Contains(err.Error(), "scalar") {
		t.Fatalf("expected nested-value error, got %v", err)
	}
}

func TestDecodeDurationsRequireNonNegativeIntegerMilliseconds(t *testing.T) {
	t.Parallel()

	if got, err := decodeMilliseconds(runtime.NewInt(25), "timeout"); err != nil || got.Milliseconds() != 25 {
		t.Fatalf("decoded duration = %v, err = %v", got, err)
	}

	for _, value := range []runtime.Value{
		runtime.NewFloat(25),
		runtime.NewInt(-1),
		runtime.NewString("25"),
	} {
		if _, err := decodeMilliseconds(value, "timeout"); err == nil {
			t.Fatalf("expected duration error for %T", value)
		}
	}
}
