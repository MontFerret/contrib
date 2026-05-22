package dom

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestHTMLElementRejectsNonStandardTextAliases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	el := new(HTMLElement)

	for _, name := range []string{"text", "html"} {
		name := name

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := runtime.NewString(name)

			if err := el.Set(ctx, key, runtime.NewString("value")); err == nil {
				t.Fatalf("expected %q assignment to fail", name)
			}

			value, err := el.Get(ctx, key)
			if err != nil {
				t.Fatalf("get %q: %v", name, err)
			}

			if value != runtime.None {
				t.Fatalf("expected %q read to return none, got %v", name, value)
			}
		})
	}
}

func TestHTMLElementSetSupportsAttributeMap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	exec := new(recordingElementEvaluator)
	el := &HTMLElement{
		attributes: newElementAttributes(exec, "node"),
	}

	attrs := runtime.NewObjectWith(map[string]runtime.Value{
		"aria-label": runtime.NewString("Continue"),
	})

	if err := el.Set(ctx, runtime.NewString("attributes"), attrs); err != nil {
		t.Fatalf("set attributes: %v", err)
	}

	if exec.evalCalls != 1 {
		t.Fatalf("expected one remote attributes write, got %d", exec.evalCalls)
	}
}
