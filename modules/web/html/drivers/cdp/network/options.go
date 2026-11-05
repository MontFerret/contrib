package network

import (
	"regexp"

	"github.com/mafredri/cdp/protocol/page"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

type (
	Cookies map[string]*drivers.HTTPCookies

	Filter struct {
		Patterns []drivers.ResourceFilter
		// AllowedOrigins restricts requests to exact HTTP(S) origins. Nil leaves
		// origins unrestricted; a non-nil empty slice denies every network origin.
		AllowedOrigins []string
	}

	Options struct {
		Cookies Cookies
		Headers *drivers.HTTPHeaders
		Filter  *Filter
	}

	WaitEventOptions struct {
		URL     *regexp.Regexp
		FrameID page.FrameID
	}
)
