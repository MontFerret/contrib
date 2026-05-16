package input

import (
	"strings"

	cdpinput "github.com/mafredri/cdp/protocol/input"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func toProtocolMouseButton(value string) (cdpinput.MouseButton, error) {
	switch strings.ToLower(value) {
	case "", "left":
		return cdpinput.MouseButtonLeft, nil
	case "middle":
		return cdpinput.MouseButtonMiddle, nil
	case "right":
		return cdpinput.MouseButtonRight, nil
	default:
		return "", runtime.Errorf(runtime.ErrInvalidArgument, "unsupported mouse button: %s", value)
	}
}
