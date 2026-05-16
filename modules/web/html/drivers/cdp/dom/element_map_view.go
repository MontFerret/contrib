package dom

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type (
	elementMapView struct {
		snapshot runtime.Map
		set      elementMapSetFunc
		remove   elementMapRemoveFunc
	}

	elementMapRemoveFunc func(ctx context.Context, key runtime.Value) error

	elementMapSetFunc func(ctx context.Context, key, value runtime.Value) (runtime.Value, bool, error)
)

func newElementMapView(snapshot runtime.Map, set elementMapSetFunc, remove elementMapRemoveFunc) *elementMapView {
	if snapshot == nil {
		snapshot = runtime.NewObject()
	}

	return &elementMapView{
		snapshot: snapshot,
		set:      set,
		remove:   remove,
	}
}

func (view *elementMapView) ObjectLike() {}

func (view *elementMapView) String() string {
	return view.snapshot.String()
}

func (view *elementMapView) Hash() uint64 {
	return view.snapshot.Hash()
}

func (view *elementMapView) Copy() runtime.Value {
	return view.snapshot.Copy()
}

func (view *elementMapView) Clone(ctx context.Context) (runtime.Cloneable, error) {
	return view.snapshot.Clone(ctx)
}

func (view *elementMapView) Compare(other runtime.Value) int {
	return view.snapshot.Compare(other)
}

func (view *elementMapView) Length(ctx context.Context) (runtime.Int, error) {
	return view.snapshot.Length(ctx)
}

func (view *elementMapView) IsEmpty(ctx context.Context) (runtime.Boolean, error) {
	length, err := view.snapshot.Length(ctx)
	if err != nil {
		return runtime.False, err
	}

	return length == 0, nil
}

func (view *elementMapView) Iterate(ctx context.Context) (runtime.Iterator, error) {
	return view.snapshot.Iterate(ctx)
}

func (view *elementMapView) Contains(ctx context.Context, value runtime.Value) (runtime.Boolean, error) {
	return view.snapshot.Contains(ctx, value)
}

func (view *elementMapView) Get(ctx context.Context, key runtime.Value) (runtime.Value, error) {
	return view.snapshot.Get(ctx, key)
}

func (view *elementMapView) Lookup(ctx context.Context, key runtime.Value) (runtime.Value, bool, error) {
	return view.snapshot.Lookup(ctx, key)
}

func (view *elementMapView) Set(ctx context.Context, key, value runtime.Value) error {
	if value == nil {
		value = runtime.None
	}

	snapshotValue, remove, err := view.set(ctx, key, value)
	if err != nil {
		return err
	}

	if remove {
		return view.snapshot.RemoveKey(ctx, key)
	}

	return view.snapshot.Set(ctx, key, snapshotValue)
}

func (view *elementMapView) RemoveKey(ctx context.Context, key runtime.Value) error {
	if err := view.remove(ctx, key); err != nil {
		return err
	}

	return view.snapshot.RemoveKey(ctx, key)
}

func (view *elementMapView) Remove(ctx context.Context, value runtime.Value) error {
	var foundKey runtime.Value

	err := view.snapshot.ForEach(ctx, func(ctx context.Context, current, key runtime.Value) (runtime.Boolean, error) {
		if runtime.CompareValues(current, value) == 0 {
			foundKey = key

			return runtime.False, nil
		}

		return runtime.True, nil
	})
	if err != nil || foundKey == nil {
		return err
	}

	return view.RemoveKey(ctx, foundKey)
}

func (view *elementMapView) Clear(ctx context.Context) error {
	keys, err := view.snapshot.Keys(ctx)
	if err != nil {
		return err
	}

	return keys.ForEach(ctx, func(ctx context.Context, key runtime.Value, _ runtime.Int) (runtime.Boolean, error) {
		if err := view.RemoveKey(ctx, key); err != nil {
			return runtime.False, err
		}

		return runtime.True, nil
	})
}

func (view *elementMapView) Empty(_ context.Context) (runtime.Map, error) {
	return runtime.NewObject(), nil
}

func (view *elementMapView) Merge(ctx context.Context, other runtime.Map) error {
	return other.ForEach(ctx, func(ctx context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		if err := view.Set(ctx, key, value); err != nil {
			return runtime.False, err
		}

		return runtime.True, nil
	})
}

func (view *elementMapView) ContainsKey(ctx context.Context, key runtime.Value) (runtime.Boolean, error) {
	return view.snapshot.ContainsKey(ctx, key)
}

func (view *elementMapView) Keys(ctx context.Context) (runtime.List, error) {
	return view.snapshot.Keys(ctx)
}

func (view *elementMapView) Values(ctx context.Context) (runtime.List, error) {
	return view.snapshot.Values(ctx)
}

func (view *elementMapView) Filter(ctx context.Context, predicate runtime.KeyReadablePredicate) (runtime.List, error) {
	return view.snapshot.Filter(ctx, predicate)
}

func (view *elementMapView) Find(ctx context.Context, predicate runtime.KeyReadablePredicate) (runtime.Value, runtime.Boolean, error) {
	return view.snapshot.Find(ctx, predicate)
}

func (view *elementMapView) ForEach(ctx context.Context, predicate runtime.KeyReadablePredicate) error {
	return view.snapshot.ForEach(ctx, predicate)
}
