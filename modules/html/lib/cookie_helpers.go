package html

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func parseCookie(ctx context.Context, value runtime.Value) (drivers.HTTPCookie, error) {
	if value == nil {
		return drivers.HTTPCookie{}, fmt.Errorf("cookie is required")
	}

	switch v := value.(type) {
	case drivers.HTTPCookie:
		return v, nil
	case *drivers.HTTPCookie:
		if v == nil {
			return drivers.HTTPCookie{}, fmt.Errorf("cookie is required")
		}

		return *v, nil
	}

	if cookie, ok := runtime.UnwrapAs[drivers.HTTPCookie](value); ok {
		return cookie, nil
	}

	if cookie, ok := runtime.UnwrapAs[*drivers.HTTPCookie](value); ok {
		if cookie == nil {
			return drivers.HTTPCookie{}, fmt.Errorf("cookie is required")
		}

		return *cookie, nil
	}

	m, err := runtime.CastMap(value)
	if err != nil {
		return drivers.HTTPCookie{}, err
	}

	name, err := getRequiredString(ctx, m, "name")
	if err != nil {
		return drivers.HTTPCookie{}, err
	}

	val, err := getRequiredString(ctx, m, "value")
	if err != nil {
		return drivers.HTTPCookie{}, err
	}

	cookie := drivers.HTTPCookie{
		Name:     name,
		Value:    val,
		SameSite: drivers.SameSiteDefaultMode,
	}

	if path, ok, err := getOptionalString(ctx, m, "path"); err != nil {
		return drivers.HTTPCookie{}, err
	} else if ok {
		cookie.Path = path
	}

	if domain, ok, err := getOptionalString(ctx, m, "domain"); err != nil {
		return drivers.HTTPCookie{}, err
	} else if ok {
		cookie.Domain = domain
	}

	if maxAge, ok, err := getOptionalInt(ctx, m, "maxAge"); err != nil {
		return drivers.HTTPCookie{}, err
	} else if ok {
		cookie.MaxAge = int(maxAge)
	}

	if expires, ok, err := getOptionalDateTime(ctx, m, "expires"); err != nil {
		return drivers.HTTPCookie{}, err
	} else if ok {
		cookie.Expires = expires
	}

	if sameSite, ok, err := getOptionalString(ctx, m, "sameSite"); err != nil {
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

	if httpOnly, ok, err := getOptionalBool(ctx, m, "httpOnly"); err != nil {
		return drivers.HTTPCookie{}, err
	} else if ok {
		cookie.HTTPOnly = httpOnly
	}

	if secure, ok, err := getOptionalBool(ctx, m, "secure"); err != nil {
		return drivers.HTTPCookie{}, err
	} else if ok {
		cookie.Secure = secure
	}

	return cookie, nil
}

func getRequiredString(ctx context.Context, m runtime.Map, key string) (string, error) {
	val, ok, err := sdk.TryGetByKey[runtime.Value](ctx, m, runtime.String(key))
	if err != nil {
		return "", err
	}

	if !ok {
		return "", fmt.Errorf("cookie %s is required", key)
	}

	if err := runtime.ValidateType(val, runtime.TypeString); err != nil {
		return "", err
	}

	return val.String(), nil
}

func getOptionalString(ctx context.Context, m runtime.Map, key string) (string, bool, error) {
	val, ok, err := sdk.TryGetByKey[runtime.Value](ctx, m, runtime.String(key))
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

func getOptionalInt(ctx context.Context, m runtime.Map, key string) (runtime.Int, bool, error) {
	val, ok, err := sdk.TryGetByKey[runtime.Value](ctx, m, runtime.String(key))
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

func getOptionalBool(ctx context.Context, m runtime.Map, key string) (bool, bool, error) {
	val, ok, err := sdk.TryGetByKey[runtime.Value](ctx, m, runtime.String(key))
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

func getOptionalDateTime(ctx context.Context, m runtime.Map, key string) (time.Time, bool, error) {
	val, ok, err := sdk.TryGetByKey[runtime.Value](ctx, m, runtime.String(key))
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
