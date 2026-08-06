package dom

import (
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

var (
	_ runtime.Comparable = (*elementMapView)(nil)
	_ runtime.Equatable  = (*elementMapView)(nil)
)

func valuesEqual(t *testing.T, actual, expected runtime.Value) bool {
	t.Helper()

	equal, err := runtime.EqualValues(t.Context(), actual, expected)
	if err != nil {
		t.Fatalf("compare runtime values: %v", err)
	}

	return bool(equal)
}
