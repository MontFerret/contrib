package lib

import (
	"context"
	"regexp"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Frames finds HTML frames by a given property selector.
// Returns an empty array if frames not found.
// @param {HTMLPage} page - HTML page.
// @param {String} property - Property selector.
// @param {String} exp - Regular expression to match property value.
// @return {HTMLDocument[]} - Returns an array of found HTML frames.
func Frames(ctx context.Context, arg1, arg2, arg3 runtime.Value) (runtime.Value, error) {
	page, err := drivers.ToPage(arg1)

	if err != nil {
		return runtime.None, runtime.ArgError(err, 0)
	}

	frames, err := page.GetFrames(ctx)

	if err != nil {
		return runtime.None, err
	}

	propName := runtime.ToString(arg2)
	matcher, err := regexp.Compile(runtime.ToString(arg3).String())

	if err != nil {
		return runtime.None, err
	}

	return frames.Filter(ctx, func(ctx context.Context, value runtime.Value, _ runtime.Int) (runtime.Boolean, error) {
		doc, e := drivers.ToDocument(value)

		if e != nil {
			err = e
			return false, e
		}

		currentPropValue, e := doc.Get(ctx, propName)

		if e != nil {
			err = e

			return false, e
		}

		return runtime.Boolean(matcher.MatchString(currentPropValue.String())), nil
	})
}
