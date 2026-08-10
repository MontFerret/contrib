package drivers

import (
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const (
	InitScriptAfterNavigation InitScriptTiming = "afterNavigation"
	InitScriptBeforeDocument  InitScriptTiming = "beforeDocument"
)

// NormalizeInitScript validates an init script and applies its default timing.
func NormalizeInitScript(script *InitScript) (*InitScript, error) {
	if script == nil {
		return nil, nil
	}

	result := *script
	if strings.TrimSpace(result.Source) == "" {
		return nil, runtime.Error(runtime.ErrInvalidArgument, "initScript source must not be empty")
	}

	if result.Timing == "" {
		result.Timing = InitScriptAfterNavigation
	}

	switch result.Timing {
	case InitScriptAfterNavigation, InitScriptBeforeDocument:
		return &result, nil
	default:
		return nil, runtime.Errorf(runtime.ErrInvalidArgument, "unsupported initScript timing: %s", result.Timing)
	}
}
