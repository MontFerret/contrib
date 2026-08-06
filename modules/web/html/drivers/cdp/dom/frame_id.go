package dom

import (
	"context"
	"strings"

	"github.com/goccy/go-json"
	"github.com/mafredri/cdp/protocol/page"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

var FrameIDType = runtime.NewTypeFor[FrameID]()

type FrameID page.FrameID

func NewFrameID(id page.FrameID) FrameID {
	return FrameID(id)
}

func (f FrameID) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(f))
}

func (f FrameID) Type() runtime.Type {
	return FrameIDType
}

func (f FrameID) String() string {
	return string(f)
}

func (f FrameID) Equal(ctx context.Context, other runtime.Value) (bool, error) {
	if _, err := checkComparisonContext(ctx); err != nil {
		return false, err
	}

	otherFrame, ok := other.(FrameID)
	if !ok {
		return false, nil
	}

	comparison, err := f.compare(ctx, otherFrame)

	return comparison == runtime.Equal, err
}

func (f FrameID) Compare(ctx context.Context, other runtime.Value) (runtime.Ordering, error) {
	if comparison, err := checkComparisonContext(ctx); err != nil {
		return comparison, err
	}

	otherFrame, ok := other.(FrameID)
	if !ok {
		return invalidComparison(f, other)
	}

	return f.compare(ctx, otherFrame)
}

func (f FrameID) compare(ctx context.Context, other FrameID) (runtime.Ordering, error) {
	if comparison, err := checkComparisonContext(ctx); err != nil {
		return comparison, err
	}

	return runtime.Ordering(strings.Compare(string(f), string(other))), nil
}

func (f FrameID) Unwrap() any {
	return page.FrameID(f)
}

func (f FrameID) Hash() uint64 {
	return runtime.Hash(FrameIDType.String(), []byte(f))
}

func (f FrameID) Copy() runtime.Value {
	return f
}
