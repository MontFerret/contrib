package dom

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type failingEqualityValue struct {
	err error
}

func newFailingEqualityValue(err error) failingEqualityValue {
	return failingEqualityValue{err: err}
}

func (value failingEqualityValue) String() string {
	return "failing equality value"
}

func (value failingEqualityValue) Hash() uint64 {
	return 1
}

func (value failingEqualityValue) Copy() runtime.Value {
	return value
}

func (value failingEqualityValue) Equal(context.Context, runtime.Value) (bool, error) {
	return false, value.err
}
