package eval

import (
	"testing"

	ferretruntime "github.com/MontFerret/ferret/v2/pkg/runtime"
)

func valuesEqual(t *testing.T, actual, expected ferretruntime.Value) bool {
	t.Helper()

	equal, err := ferretruntime.EqualValues(t.Context(), actual, expected)
	if err != nil {
		t.Fatalf("compare runtime values: %v", err)
	}

	return bool(equal)
}
