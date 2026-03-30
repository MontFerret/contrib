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
