package dom

import (
	"context"
	"testing"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

var (
	_ drivers.AttributeTarget = (*elementAttributes)(nil)
	_ drivers.StyleTarget     = (*elementStyles)(nil)
	_ drivers.WaitTarget      = (*elementWait)(nil)
)

func TestAttributeViewWritesThroughCapabilityAndSnapshot(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	exec := &recordingElementEvaluator{
		value: runtime.NewObjectWith(map[string]runtime.Value{
			"role": runtime.NewString("hero"),
		}),
	}

	view, err := newAttributeView(ctx, newElementAttributes(exec, "node"))
	if err != nil {
		t.Fatalf("new attribute view: %v", err)
	}

	if err := view.Set(ctx, runtime.NewString("role"), runtime.NewString("banner")); err != nil {
		t.Fatalf("set role: %v", err)
	}

	assertViewValue(t, ctx, view, "role", runtime.NewString("banner"))
	if exec.evalCalls != 1 {
		t.Fatalf("expected one remote write, got %d", exec.evalCalls)
	}

	if err := view.RemoveKey(ctx, runtime.NewString("role")); err != nil {
		t.Fatalf("remove role: %v", err)
	}

	assertViewValue(t, ctx, view, "role", runtime.None)
	if exec.evalCalls != 2 {
		t.Fatalf("expected two remote writes, got %d", exec.evalCalls)
	}
}

func TestStyleViewWritesThroughCapabilityAndSnapshot(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	exec := &recordingElementEvaluator{
		value: runtime.NewObjectWith(map[string]runtime.Value{
			"display": runtime.NewString("none"),
		}),
	}

	view, err := newStyleView(ctx, newElementStyles(exec, "node"))
	if err != nil {
		t.Fatalf("new style view: %v", err)
	}

	if err := view.Set(ctx, runtime.NewString("display"), runtime.NewString("block")); err != nil {
		t.Fatalf("set display: %v", err)
	}

	assertViewValue(t, ctx, view, "display", runtime.NewString("block"))
	if exec.evalCalls != 1 {
		t.Fatalf("expected one remote write, got %d", exec.evalCalls)
	}

	if err := view.Set(ctx, runtime.NewString("display"), runtime.None); err != nil {
		t.Fatalf("remove display via none: %v", err)
	}

	assertViewValue(t, ctx, view, "display", runtime.None)
	if exec.evalCalls != 2 {
		t.Fatalf("expected two remote writes, got %d", exec.evalCalls)
	}
}

func TestStyleViewNormalizesSnapshotKeys(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	exec := &recordingElementEvaluator{
		value: runtime.NewObjectWith(map[string]runtime.Value{
			"background-color": runtime.NewString("transparent"),
		}),
	}

	view, err := newStyleView(ctx, newElementStyles(exec, "node"))
	if err != nil {
		t.Fatalf("new style view: %v", err)
	}

	assertViewValue(t, ctx, view, "backgroundColor", runtime.NewString("transparent"))

	if err := view.Set(ctx, runtime.NewString("backgroundColor"), runtime.NewString("red")); err != nil {
		t.Fatalf("set backgroundColor: %v", err)
	}

	assertViewValue(t, ctx, view, "background-color", runtime.NewString("red"))
	assertViewValue(t, ctx, view, "backgroundColor", runtime.NewString("red"))

	if err := view.Set(ctx, runtime.NewString("backgroundColor"), runtime.None); err != nil {
		t.Fatalf("remove backgroundColor: %v", err)
	}

	assertViewValue(t, ctx, view, "background-color", runtime.None)
	if exec.evalCalls != 2 {
		t.Fatalf("expected two remote writes, got %d", exec.evalCalls)
	}
}

func TestDatasetViewNormalizesSnapshotKeys(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	exec := &recordingElementEvaluator{
		value: runtime.NewObjectWith(map[string]runtime.Value{
			"productId": runtime.NewString("old"),
			"foo_bar":   runtime.NewString("kept"),
		}),
	}

	view, err := newDatasetView(ctx, newElementDataset(exec, "node"))
	if err != nil {
		t.Fatalf("new dataset view: %v", err)
	}

	assertViewValue(t, ctx, view, "data-product-id", runtime.NewString("old"))

	if err := view.Set(ctx, runtime.NewString("data-product-id"), runtime.NewString("123")); err != nil {
		t.Fatalf("set data-product-id: %v", err)
	}

	assertViewValue(t, ctx, view, "productId", runtime.NewString("123"))
	assertViewValue(t, ctx, view, "data-product-id", runtime.NewString("123"))

	if err := view.Set(ctx, runtime.NewString("data-foo_bar"), runtime.NewString("stable")); err != nil {
		t.Fatalf("set data-foo_bar: %v", err)
	}

	assertViewValue(t, ctx, view, "foo_bar", runtime.NewString("stable"))
	assertViewValue(t, ctx, view, "fooBar", runtime.None)

	if err := view.Set(ctx, runtime.NewString("data-product-id"), runtime.None); err != nil {
		t.Fatalf("remove data-product-id: %v", err)
	}

	assertViewValue(t, ctx, view, "productId", runtime.None)
	if exec.evalCalls != 3 {
		t.Fatalf("expected three remote writes, got %d", exec.evalCalls)
	}
}

func TestClassListViewWritesThroughCapabilityAndSnapshot(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	exec := &recordingElementEvaluator{
		value: runtime.NewObjectWith(map[string]runtime.Value{
			"active": runtime.True,
		}),
	}

	view, err := newClassListView(ctx, newElementClasses(exec, "node"))
	if err != nil {
		t.Fatalf("new class list view: %v", err)
	}

	if err := view.Set(ctx, runtime.NewString("active"), runtime.False); err != nil {
		t.Fatalf("disable active: %v", err)
	}

	assertViewValue(t, ctx, view, "active", runtime.None)
	if exec.evalCalls != 1 {
		t.Fatalf("expected one remote write, got %d", exec.evalCalls)
	}

	if err := view.Set(ctx, runtime.NewString("fresh"), runtime.True); err != nil {
		t.Fatalf("enable fresh: %v", err)
	}

	assertViewValue(t, ctx, view, "fresh", runtime.True)
	if exec.evalCalls != 2 {
		t.Fatalf("expected two remote writes, got %d", exec.evalCalls)
	}
}

func assertViewValue(t *testing.T, ctx context.Context, view runtime.KeyReadable, key string, want runtime.Value) {
	t.Helper()

	got, err := view.Get(ctx, runtime.NewString(key))
	if err != nil {
		t.Fatalf("get %s: %v", key, err)
	}

	if runtime.CompareValues(got, want) != 0 {
		t.Fatalf("unexpected %s value: got %v, want %v", key, got, want)
	}
}
