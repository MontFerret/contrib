package dom

import (
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDatasetPropertyName(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"productId":       "productId",
		"product-id":      "productId",
		"data-product-id": "productId",
		"data-foo_bar":    "foo_bar",
	}

	for input, want := range tests {
		input, want := input, want
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			got := datasetPropertyName(runtime.NewString(input))
			if got.String() != want {
				t.Fatalf("expected %q, got %q", want, got.String())
			}
		})
	}
}

func TestStylePropertyName(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"display":             "display",
		"backgroundColor":     "background-color",
		"borderTopLeftRadius": "border-top-left-radius",
		"background-color":    "background-color",
		"--accent-color":      "--accent-color",
	}

	for input, want := range tests {
		input, want := input, want
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			got := stylePropertyName(runtime.NewString(input))
			if got.String() != want {
				t.Fatalf("expected %q, got %q", want, got.String())
			}
		})
	}
}
