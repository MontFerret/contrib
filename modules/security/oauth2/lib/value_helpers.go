package lib

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

var hostValueIdentity atomic.Uint64

func nextHostValueIdentity() uint64 {
	return hostValueIdentity.Add(1)
}

func safeKey(key runtime.Value) string {
	if key == nil || key == runtime.None {
		return ""
	}

	return key.String()
}

func stringArray(values []string) *runtime.Array {
	items := make([]runtime.Value, len(values))
	for i, value := range values {
		items[i] = runtime.NewString(value)
	}

	return runtime.NewArrayOf(items)
}

func stringValue(value string) runtime.Value {
	if value == "" {
		return runtime.None
	}

	return runtime.NewString(value)
}

func lookupValue(ctx context.Context, value runtime.KeyReadable, key string) (runtime.Value, error) {
	return value.Get(ctx, runtime.NewString(key))
}

func ceilMilliseconds(value time.Duration) int64 {
	if value <= 0 {
		return 0
	}

	milliseconds := value / time.Millisecond
	if value%time.Millisecond != 0 {
		milliseconds++
	}

	return int64(milliseconds)
}
