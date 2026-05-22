package dom

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestElementMapViewWritesThroughAndPreservesSnapshotSemantics(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	writes := make(map[string]runtime.Value)
	removes := make(map[string]bool)
	view := newElementMapView(
		runtime.NewObjectWith(map[string]runtime.Value{
			"stale": runtime.NewString("snapshot"),
			"drop":  runtime.NewString("remove"),
		}),
		func(_ context.Context, key, value runtime.Value) (runtime.Value, bool, error) {
			writes[key.String()] = value
			if value == runtime.None {
				return runtime.None, true, nil
			}

			return runtime.ToString(value), false, nil
		},
		func(_ context.Context, key runtime.Value) error {
			removes[key.String()] = true

			return nil
		},
	)

	if err := view.Set(ctx, runtime.NewString("fresh"), runtime.NewInt(42)); err != nil {
		t.Fatalf("set fresh: %v", err)
	}

	if got, err := view.Get(ctx, runtime.NewString("fresh")); err != nil {
		t.Fatalf("get fresh: %v", err)
	} else if runtime.CompareValues(got, runtime.NewString("42")) != 0 {
		t.Fatalf("unexpected fresh snapshot value: %v", got)
	}

	if _, ok := writes["fresh"]; !ok {
		t.Fatal("expected fresh write-through")
	}

	if err := view.Set(ctx, runtime.NewString("drop"), runtime.None); err != nil {
		t.Fatalf("remove via none: %v", err)
	}

	if got, err := view.Get(ctx, runtime.NewString("drop")); err != nil {
		t.Fatalf("get dropped key: %v", err)
	} else if got != runtime.None {
		t.Fatalf("expected dropped snapshot key to be none, got %v", got)
	}

	merge := runtime.NewObjectWith(map[string]runtime.Value{
		"merged": runtime.NewString("value"),
	})
	if err := view.Merge(ctx, merge); err != nil {
		t.Fatalf("merge: %v", err)
	}

	if _, ok := writes["merged"]; !ok {
		t.Fatal("expected merged key to write through")
	}

	if err := view.Clear(ctx); err != nil {
		t.Fatalf("clear: %v", err)
	}

	for _, key := range []string{"stale", "fresh", "merged"} {
		if !removes[key] {
			t.Fatalf("expected clear to remove %q", key)
		}
	}
}

func TestElementMapViewNormalizesSnapshotAndWriteKeys(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	writes := make(map[string]runtime.Value)
	removes := make(map[string]bool)
	view := newElementMapView(
		runtime.NewObjectWith(map[string]runtime.Value{
			"canonical": runtime.NewString("snapshot"),
		}),
		func(_ context.Context, key, value runtime.Value) (runtime.Value, bool, error) {
			writes[key.String()] = value
			if value == runtime.None {
				return runtime.None, true, nil
			}

			return runtime.ToString(value), false, nil
		},
		func(_ context.Context, key runtime.Value) error {
			removes[key.String()] = true

			return nil
		},
	).withKeyNormalizer(func(key runtime.Value) runtime.Value {
		if key.String() == "alias" {
			return runtime.NewString("canonical")
		}

		return key
	})

	assertViewValue(t, ctx, view, "alias", runtime.NewString("snapshot"))

	if err := view.Set(ctx, runtime.NewString("alias"), runtime.NewString("fresh")); err != nil {
		t.Fatalf("set alias: %v", err)
	}

	assertViewValue(t, ctx, view, "canonical", runtime.NewString("fresh"))
	assertViewValue(t, ctx, view, "alias", runtime.NewString("fresh"))
	if _, ok := writes["canonical"]; !ok {
		t.Fatal("expected write-through to use normalized key")
	}
	if _, ok := writes["alias"]; ok {
		t.Fatal("did not expect write-through to use raw alias key")
	}

	if err := view.RemoveKey(ctx, runtime.NewString("alias")); err != nil {
		t.Fatalf("remove alias: %v", err)
	}

	assertViewValue(t, ctx, view, "canonical", runtime.None)
	if !removes["canonical"] {
		t.Fatal("expected remove to use normalized key")
	}
	if removes["alias"] {
		t.Fatal("did not expect remove to use raw alias key")
	}
}
