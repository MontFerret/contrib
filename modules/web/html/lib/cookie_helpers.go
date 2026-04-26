package lib

import (
	"context"
	"strings"
	"time"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func parseCookie(ctx context.Context, value runtime.Value) (drivers.HTTPCookie, error) {
	if value == nil {
		return drivers.HTTPCookie{}, runtime.Error(runtime.ErrMissedArgument, "cookie")
	}

	switch v := value.(type) {
	case drivers.HTTPCookie:
		return v, nil
	case *drivers.HTTPCookie:
		if v == nil {
			return drivers.HTTPCookie{}, runtime.Error(runtime.ErrMissedArgument, "cookie")
		}

		return *v, nil
	}

	if cookie, ok := runtime.UnwrapAs[drivers.HTTPCookie](value); ok {
		return cookie, nil
	}

	if cookie, ok := runtime.UnwrapAs[*drivers.HTTPCookie](value); ok {
		if cookie == nil {
			return drivers.HTTPCookie{}, runtime.Error(runtime.ErrMissedArgument, "cookie")
		}

		return *cookie, nil
	}

	m, err := runtime.CastMap(value)
	if err != nil {
		return drivers.HTTPCookie{}, err
	}

	name, err := getRequiredString(ctx, m, "name", "Name")
	if err != nil {
		return drivers.HTTPCookie{}, err
	}

	val, err := getRequiredString(ctx, m, "value", "Value")
	if err != nil {
		return drivers.HTTPCookie{}, err
	}

	cookie := drivers.HTTPCookie{
		Name:     name,
		Value:    val,
		SameSite: drivers.SameSiteDefaultMode,
	}

	if path, ok, err := getOptionalString(ctx, m, "path", "Path"); err != nil {
		return drivers.HTTPCookie{}, err
	} else if ok {
		cookie.Path = path
	}

	if domain, ok, err := getOptionalString(ctx, m, "domain", "Domain"); err != nil {
		return drivers.HTTPCookie{}, err
	} else if ok {
		cookie.Domain = domain
	}

	if maxAge, ok, err := getOptionalInt(ctx, m, "maxAge", "MaxAge", "max_age"); err != nil {
		return drivers.HTTPCookie{}, err
	} else if ok {
		cookie.MaxAge = int(maxAge)
	}

	if expires, ok, err := getOptionalDateTime(ctx, m, "expires", "Expires"); err != nil {
		return drivers.HTTPCookie{}, err
	} else if ok {
		cookie.Expires = expires
	}

	if sameSite, ok, err := getOptionalString(ctx, m, "sameSite", "SameSite", "same_site"); err != nil {
		return drivers.HTTPCookie{}, err
	} else if ok {
		switch strings.ToLower(sameSite) {
		case "lax":
			cookie.SameSite = drivers.SameSiteLaxMode
		case "strict":
			cookie.SameSite = drivers.SameSiteStrictMode
		default:
			cookie.SameSite = drivers.SameSiteDefaultMode
		}
	}

	if httpOnly, ok, err := getOptionalBool(ctx, m, "httpOnly", "HTTPOnly", "http_only"); err != nil {
		return drivers.HTTPCookie{}, err
	} else if ok {
		cookie.HTTPOnly = httpOnly
	}

	if secure, ok, err := getOptionalBool(ctx, m, "secure", "Secure"); err != nil {
		return drivers.HTTPCookie{}, err
	} else if ok {
		cookie.Secure = secure
	}

	return cookie, nil
}

func parseCookiesValue(ctx context.Context, value runtime.Value) (*drivers.HTTPCookies, error) {
	if value == nil || value == runtime.None {
		return nil, runtime.Error(runtime.ErrMissedArgument, "cookies")
	}

	switch v := value.(type) {
	case *drivers.HTTPCookies:
		if v == nil {
			return nil, runtime.Error(runtime.ErrMissedArgument, "cookies")
		}

		cookies := drivers.NewHTTPCookies()

		for _, cookie := range v.Data {
			cookies.SetCookie(cookie)
		}

		return cookies, nil
	}

	if cookies, ok := runtime.UnwrapAs[*drivers.HTTPCookies](value); ok {
		if cookies == nil {
			return nil, runtime.Error(runtime.ErrMissedArgument, "cookies")
		}

		res := drivers.NewHTTPCookies()

		for _, cookie := range cookies.Data {
			res.SetCookie(cookie)
		}

		return res, nil
	}

	if list, err := runtime.CastList(value); err == nil {
		cookies := drivers.NewHTTPCookies()

		err = list.ForEach(ctx, func(ctx context.Context, value runtime.Value, _ runtime.Int) (runtime.Boolean, error) {
			cookie, err := parseCookie(ctx, value)
			if err != nil {
				return false, err
			}

			cookies.SetCookie(cookie)

			return true, nil
		})

		if err != nil {
			return nil, err
		}

		return cookies, nil
	}

	cookie, err := parseCookie(ctx, value)
	if err != nil {
		return nil, err
	}

	cookies := drivers.NewHTTPCookies()
	cookies.SetCookie(cookie)

	return cookies, nil
}

func getRequiredString(ctx context.Context, m runtime.Map, key string, aliases ...string) (string, error) {
	val, ok, err := getCookieField(ctx, m, key, aliases...)
	if err != nil {
		return "", err
	}

	if !ok {
		return "", runtime.Errorf(runtime.ErrMissedArgument, "cookie %s", key)
	}

	if err := runtime.ValidateType(val, runtime.TypeString); err != nil {
		return "", err
	}

	return val.String(), nil
}

func getOptionalString(ctx context.Context, m runtime.Map, key string, aliases ...string) (string, bool, error) {
	val, ok, err := getCookieField(ctx, m, key, aliases...)
	if err != nil {
		return "", false, err
	}

	if !ok {
		return "", false, nil
	}

	if err := runtime.ValidateType(val, runtime.TypeString); err != nil {
		return "", false, err
	}

	return val.String(), true, nil
}

func getOptionalInt(ctx context.Context, m runtime.Map, key string, aliases ...string) (runtime.Int, bool, error) {
	val, ok, err := getCookieField(ctx, m, key, aliases...)
	if err != nil {
		return 0, false, err
	}

	if !ok {
		return 0, false, nil
	}

	if err := runtime.ValidateType(val, runtime.TypeInt); err != nil {
		return 0, false, err
	}

	return val.(runtime.Int), true, nil
}

func getOptionalBool(ctx context.Context, m runtime.Map, key string, aliases ...string) (bool, bool, error) {
	val, ok, err := getCookieField(ctx, m, key, aliases...)
	if err != nil {
		return false, false, err
	}

	if !ok {
		return false, false, nil
	}

	if err := runtime.ValidateType(val, runtime.TypeBoolean); err != nil {
		return false, false, err
	}

	return bool(val.(runtime.Boolean)), true, nil
}

func getOptionalDateTime(ctx context.Context, m runtime.Map, key string, aliases ...string) (time.Time, bool, error) {
	val, ok, err := getCookieField(ctx, m, key, aliases...)
	if err != nil {
		return time.Time{}, false, err
	}

	if !ok {
		return time.Time{}, false, nil
	}

	if err := runtime.ValidateType(val, runtime.TypeDateTime, runtime.TypeString); err != nil {
		return time.Time{}, false, err
	}

	if dt, ok := val.(runtime.DateTime); ok {
		return dt.Unwrap().(time.Time), true, nil
	}

	parsed, err := time.Parse(runtime.DefaultTimeLayout, val.String())
	if err != nil {
		return time.Time{}, false, err
	}

	return parsed, true, nil
}

func getCookieField(ctx context.Context, m runtime.Map, key string, aliases ...string) (runtime.Value, bool, error) {
	keys := append([]string{key}, aliases...)

	for _, k := range keys {
		val, ok, err := sdk.TryGetByKey[runtime.Value](ctx, m, runtime.String(k))
		if err != nil {
			return runtime.None, false, err
		}

		if ok {
			return val, true, nil
		}
	}

	return runtime.None, false, nil
}
