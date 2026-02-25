package html

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// COOKIE_GET gets a cookie from a given page by name.
// @param {HTMLPage} page - Target page.
// @param {String} name - Cookie or cookie name to delete.
// @return {HTTPCookie} - Cookie if found, otherwise None.
func CookieGet(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	//err := runtime.ValidateArgs(args, 2, 2)
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
	//err = runtime.ValidateType(args[1], runtime.TypeString)
	//
	//if err != nil {
	//	return runtime.None, err
	//}
	//
	//name := args[1].(runtime.String)
	//
	//cookies, err := page.GetCookies(ctx)
	//
	//if err != nil {
	//	return runtime.None, err
	//}
	//
	//cookie, found := cookies.Get(name)
	//
	//if found {
	//	return cookie, nil
	//}
	//
	//return runtime.None, nil

	return runtime.None, nil
}
