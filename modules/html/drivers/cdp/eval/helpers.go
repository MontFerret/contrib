package eval

import (
	"strings"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
)

func parseRuntimeException(details *cdpruntime.ExceptionDetails) error {
	if details == nil || details.Exception == nil {
		return nil
	}

	desc := *details.Exception.Description

	if strings.Contains(desc, drivers.ErrNotFound.Error()) {
		return drivers.ErrNotFound
	}

	return runtime.Error(
		runtime.ErrUnexpected,
		desc,
	)
}
