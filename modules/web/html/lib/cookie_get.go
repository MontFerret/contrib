package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// CookieGet gets a cookie from a given page by name.
// @param {HTMLPage} page - Target page.
// @param {String} name - Cookie or cookie name to delete.
// @return {HTTPCookie} - Cookie if found, otherwise None.
func CookieGet(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 2, 2); err != nil {
		return runtime.None, err
	}

	page, err := runtime.CastArgAt[drivers.HTMLPage](args, 0)
	if err != nil {
		return runtime.None, err
	}

	if err := runtime.ValidateArgTypeAt(args, 1, runtime.TypeString); err != nil {
		return runtime.None, err
	}

	name := args[1].(runtime.String)
	cookies, err := page.GetCookies(ctx)
	if err != nil {
		return runtime.None, err
	}

	cookie, err := cookies.Get(ctx, name)
	if err != nil {
		return runtime.None, err
	}

	if cookie == runtime.None {
		return runtime.None, nil
	}

	return cookie, nil
}
