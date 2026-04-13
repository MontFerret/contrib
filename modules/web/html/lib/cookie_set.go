package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// CookieSet sets cookies on a given page.
// @param {HTMLPage} page - Target page.
// @param {HTTPCookie, repeated} cookies - Target cookies.
func CookieSet(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 2, runtime.MaxArgs); err != nil {
		return runtime.None, err
	}

	page, err := runtime.CastArgAt[drivers.HTMLPage](args, 0)
	if err != nil {
		return runtime.None, err
	}

	cookies := drivers.NewHTTPCookies()

	for _, c := range args[1:] {
		parsed, err := parseCookiesValue(ctx, c)
		if err != nil {
			return runtime.None, err
		}

		for _, cookie := range parsed.Data {
			cookies.SetCookie(cookie)
		}
	}

	return runtime.None, page.SetCookies(ctx, cookies)
}
