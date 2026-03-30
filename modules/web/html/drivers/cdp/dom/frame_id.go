package dom

import (
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/goccy/go-json"
	"github.com/mafredri/cdp/protocol/page"
)

var FrameIDType = runtime.NewTypeFor[FrameID]("html.drivers.cdp.dom", "FrameID")

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

func (f FrameID) Compare(other runtime.Value) int64 {
	var s1 string
	var s2 string

	s1 = string(f)

	switch v := other.(type) {
	case FrameID:
		s2 = string(v)
	case *HTMLDocument:
		s2 = string(v.Frame().Frame.ID)
	case runtime.String:
		s2 = v.String()
	default:
		return -1
	}

	return int64(strings.Compare(s1, s2))
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
