package html

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// COOKIE_DEL gets a cookie from a given page by name.
// @param {HTMLPage} page - Target page.
// @param {HTTPCookie, repeated | String, repeated} cookiesOrNames - Cookie or cookie name to delete.
func CookieDel(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	//err := runtime.ValidateArgs(args, 2, runtime.MaxArgs)
	//
	//if err != nil {
	//	return runtime.None, err
	//}
	//
	//page, err := drivers.ToPage(args[0])
	//
	//if err != nil {
	//	return runtime.None, err
	//}
	//
	//inputs := args[1:]
	//var currentCookies *drivers.HTTPCookies
	//cookies := drivers.NewHTTPCookies()
	//
	//for _, c := range inputs {
	//	switch cookie := c.(type) {
	//	case runtime.String:
	//		if currentCookies == nil {
	//			current, err := page.GetCookies(ctx)
	//
	//			if err != nil {
	//				return runtime.None, err
	//			}
	//
	//			currentCookies = current
	//		}
	//
	//		found, isFound, err := sdk.TryGetByKey[runtime.Value](ctx, currentCookies, cookie)
	//
	//		if isFound {
	//			cookies.Set(found)
	//		}
	//
	//	case drivers.HTTPCookie:
	//		cookies.SetCookie(cookie)
	//	default:
	//		return runtime.None, runtime.TypeErrorOf(c, runtime.TypeString, drivers.HTTPCookieType)
	//	}
	//}
	//
	//return runtime.None, page.DeleteCookies(ctx, cookies)

	return runtime.None, nil
}
