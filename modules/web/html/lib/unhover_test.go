package lib

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestUnhoverUsesDirectInteractionTarget(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><body><button id="cta">go</button></body></html>`)

	value, err := Unhover(context.Background(), doc.element)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != runtime.True {
		t.Fatalf("expected unhover to report success, got %v", value)
	}

	if !doc.element.unhovered {
		t.Fatal("expected unhover to be delegated to the target element")
	}
}

func TestUnhoverUsesSelectorInteractionTarget(t *testing.T) {
	t.Parallel()

	page := newTestPage(t, `<html><body><button id="cta">go</button></body></html>`)

	value, err := Unhover(context.Background(), page, runtime.NewString("#cta"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != runtime.True {
		t.Fatalf("expected unhover to report success, got %v", value)
	}

	if got := page.frame.element.unhoveredSelector; got != "#cta" {
		t.Fatalf("expected unhover selector to be delegated to the root element, got %q", got)
	}
}

func TestUnhoverValidatesArguments(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><body><button id="cta">go</button></body></html>`)
	cases := []struct {
		name string
		args []runtime.Value
	}{
		{
			name: "missing target",
		},
		{
			name: "too many arguments",
			args: []runtime.Value{doc, runtime.NewString("#cta"), runtime.NewString("extra")},
		},
		{
			name: "invalid selector",
			args: []runtime.Value{doc, runtime.NewInt(1)},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := Unhover(context.Background(), tc.args...); err == nil {
				t.Fatal("expected argument validation error")
			}
		})
	}
}
