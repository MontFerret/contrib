package html

import (
	"context"
	"regexp"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// FRAMES finds HTML frames by a given property selector.
// Returns an empty array if frames not found.
// @param {HTMLPage} page - HTML page.
// @param {String} property - Property selector.
// @param {String} exp - Regular expression to match property value.
// @return {HTMLDocument[]} - Returns an array of found HTML frames.
func Frames(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 3, 3)

	if err != nil {
		return runtime.None, err
	}

	page, err := drivers.ToPage(args[0])

	if err != nil {
		return runtime.None, err
	}

	frames, err := page.GetFrames(ctx)

	if err != nil {
		return runtime.None, err
	}

	propName := runtime.ToString(args[1])
	matcher, err := regexp.Compile(runtime.ToString(args[2]).String())

	if err != nil {
		return runtime.None, err
	}

	result, _ := frames.Find(ctx, func(ctx context.Context, value runtime.Value, _ runtime.Int) (runtime.Boolean, error) {
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

	return result, err
}
