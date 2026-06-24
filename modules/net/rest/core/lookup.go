package core

import (
	"context"
	"time"

	commonobject "github.com/MontFerret/contrib/pkg/common/object"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func requireMap(_ context.Context, value runtime.Value, owner string) (runtime.Map, error) {
	return commonobject.RequireMap(value, owner)
}

func lookupValue(ctx context.Context, obj runtime.Map, key string) (runtime.Value, bool, error) {
	return commonobject.Value(ctx, obj, key)
}

func lookupString(ctx context.Context, obj runtime.Map, key, owner string) (string, bool, error) {
	return commonobject.String(ctx, obj, key, owner)
}

func lookupDuration(ctx context.Context, obj runtime.Map, key, owner string) (time.Duration, bool, error) {
	return commonobject.MillisDuration(ctx, obj, key, owner)
}
