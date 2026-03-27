package types

import (
	"fmt"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

var Stringer = runtime.HostTypeOf((*fmt.Stringer)(nil))

func ResolveContent(input runtime.Value) (runtime.String, error) {
	switch content := input.(type) {
	case runtime.String:
		return content, nil
	case runtime.Binary:
		return runtime.NewString(content.String()), nil
	case fmt.Stringer:
		return runtime.NewString(content.String()), nil
	default:
		return runtime.EmptyString, runtime.TypeErrorOf(input, runtime.TypeString, runtime.TypeBinary, Stringer)
	}
}
