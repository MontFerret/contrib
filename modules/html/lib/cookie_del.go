package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// COOKIE_DEL gets a cookie from a given page by name.
// @param {HTMLPage} page - Target page.
// @param {HTTPCookie, repeated | String, repeated} cookiesOrNames - Cookie or cookie name to delete.
func CookieDel(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 2, runtime.MaxArgs); err != nil {
		return runtime.None, err
	}

	page, err := drivers.ToPage(args[0])
	if err != nil {
		return runtime.None, err
	}

	inputs := args[1:]
	var currentCookies *drivers.HTTPCookies
	cookies := drivers.NewHTTPCookies()

	for _, c := range inputs {
		switch cookie := c.(type) {
		case runtime.String:
			if currentCookies == nil {
				current, err := page.GetCookies(ctx)
				if err != nil {
					return runtime.None, err
				}

				currentCookies = current
			}

			found, err := currentCookies.Get(ctx, cookie)
			if err != nil {
				return runtime.None, err
			}

			if found == runtime.None {
				continue
			}

			if parsed, ok := runtime.UnwrapAs[drivers.HTTPCookie](found); ok {
				cookies.SetCookie(parsed)
				continue
			}

			if parsed, ok := runtime.UnwrapAs[*drivers.HTTPCookie](found); ok && parsed != nil {
				cookies.SetCookie(*parsed)
				continue
			}

			switch parsed := found.(type) {
			case drivers.HTTPCookie:
				cookies.SetCookie(parsed)
			case *drivers.HTTPCookie:
				if parsed != nil {
					cookies.SetCookie(*parsed)
				}
			default:
				return runtime.None, runtime.TypeErrorOf(found, drivers.HTTPCookieType)
			}
		default:
			parsed, err := parseCookie(ctx, c)
			if err != nil {
				return runtime.None, err
			}

			cookies.SetCookie(parsed)
		}
	}

	return runtime.None, page.DeleteCookies(ctx, cookies)
}
