package html

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// COOKIE_SET sets cookies to a given page
// @param {HTMLPage} page - Target page.
// @param {HTTPCookie, repeated} cookies - Target cookies.
func CookieSet(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
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
	//cookies := drivers.NewHTTPCookies()
	//
	//for _, c := range args[1:] {
	//	cookie, err := parseCookie(c)
	//
	//	if err != nil {
	//		return runtime.None, err
	//	}
	//
	//	cookies.Set(cookie)
	//}
	//
	//return runtime.None, page.SetCookies(ctx, cookies)

	return runtime.None, nil
}
