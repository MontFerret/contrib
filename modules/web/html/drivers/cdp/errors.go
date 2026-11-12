package cdp

import (
	"strings"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func invalidObservableEventNameError(eventName string) error {
	return runtime.Errorf(
		runtime.ErrInvalidOperation,
		"unknown event name: %s; supported events: %s",
		eventName,
		strings.Join(drivers.SupportedObservableEvents(), ", "),
	)
}
