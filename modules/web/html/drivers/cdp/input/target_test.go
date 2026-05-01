package input

import (
	"context"
	"errors"
	"testing"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestTargetRefResolveDirect(t *testing.T) {
	target := directTarget(cdpruntime.RemoteObjectID("direct"))

	got, err := target.resolve(func(parentID cdpruntime.RemoteObjectID, selector drivers.QuerySelector) (cdpruntime.RemoteObjectID, error) {
		t.Fatalf("resolver should not be called for direct targets")
		return "", nil
	})
	if err != nil {
		t.Fatalf("resolve direct target: %v", err)
	}

	if got != cdpruntime.RemoteObjectID("direct") {
		t.Fatalf("expected direct object id, got %q", got)
	}
}

func TestTargetRefResolveSelector(t *testing.T) {
	target := selectorTarget(cdpruntime.RemoteObjectID("parent"), drivers.NewCSSSelector(".hero"))

	got, err := target.resolve(func(parentID cdpruntime.RemoteObjectID, selector drivers.QuerySelector) (cdpruntime.RemoteObjectID, error) {
		if parentID != cdpruntime.RemoteObjectID("parent") {
			t.Fatalf("unexpected parent id %q", parentID)
		}
		if selector.String() != ".hero" {
			t.Fatalf("unexpected selector %q", selector)
		}

		return cdpruntime.RemoteObjectID("child"), nil
	})
	if err != nil {
		t.Fatalf("resolve selector target: %v", err)
	}

	if got != cdpruntime.RemoteObjectID("child") {
		t.Fatalf("expected resolved child id, got %q", got)
	}
}

func TestQuerySelectorObjectID(t *testing.T) {
	prev := evalRefBySelector
	t.Cleanup(func() {
		evalRefBySelector = prev
	})

	evalRefBySelector = func(
		_ context.Context,
		_ *eval.Runtime,
		parentID cdpruntime.RemoteObjectID,
		selector drivers.QuerySelector,
	) (cdpruntime.RemoteObject, error) {
		if parentID != cdpruntime.RemoteObjectID("page") {
			t.Fatalf("unexpected parent id %q", parentID)
		}
		if selector.String() != ".cta" {
			t.Fatalf("unexpected selector %q", selector)
		}

		objectID := cdpruntime.RemoteObjectID("button")

		return cdpruntime.RemoteObject{ObjectID: &objectID}, nil
	}

	manager := &Manager{}
	got, err := manager.querySelectorObjectID(context.Background(), cdpruntime.RemoteObjectID("page"), drivers.NewCSSSelector(".cta"))
	if err != nil {
		t.Fatalf("query selector object id: %v", err)
	}

	if got != cdpruntime.RemoteObjectID("button") {
		t.Fatalf("expected object id button, got %q", got)
	}
}

func TestQuerySelectorObjectIDReportsNotFound(t *testing.T) {
	prev := evalRefBySelector
	t.Cleanup(func() {
		evalRefBySelector = prev
	})

	evalRefBySelector = func(
		_ context.Context,
		_ *eval.Runtime,
		_ cdpruntime.RemoteObjectID,
		_ drivers.QuerySelector,
	) (cdpruntime.RemoteObject, error) {
		return cdpruntime.RemoteObject{}, nil
	}

	manager := &Manager{}
	_, err := manager.querySelectorObjectID(context.Background(), cdpruntime.RemoteObjectID("page"), drivers.NewCSSSelector(".missing"))
	if !errors.Is(err, runtime.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}
