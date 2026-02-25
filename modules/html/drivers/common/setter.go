package common

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func SetInPage(ctx context.Context, key runtime.Value, page drivers.HTMLPage, value runtime.Value) error {
	return SetInDocument(ctx, key, page.GetMainFrame(), value)
}

func SetInDocument(ctx context.Context, key runtime.Value, doc drivers.HTMLDocument, value runtime.Value) error {
	return SetInNode(ctx, key, doc, value)
}

func SetInElement(ctx context.Context, key runtime.Value, el drivers.HTMLElement, value runtime.Value) error {
	if IsEmptyValue(value) {
		return nil
	}

	switch key.String() {
	case "attributes":
		obj, err := runtime.CastMap(value)

		if err != nil {
			return err
		}

		curr, err := el.GetAttributes(ctx)

		if err != nil {
			return err
		}

		keys, err := curr.Keys(ctx)

		if err != nil {
			return err
		}

		keySlice, err := ToRuntimeStringSlice(ctx, keys)

		if err != nil {
			return err
		}

		// remove all previous attributes
		err = el.RemoveAttribute(ctx, keySlice...)

		if err != nil {
			return err
		}

		return obj.ForEach(ctx, func(ctx context.Context, value, key runtime.Value) (runtime.Boolean, error) {
			err = el.SetAttribute(ctx, runtime.NewString(key.String()), runtime.NewString(value.String()))

			if err != nil {
				return false, err
			}

			return true, nil
		})
	case "style":
		obj, err := runtime.CastMap(value)

		if err != nil {
			return err
		}

		curr, err := el.GetStyles(ctx)

		if err != nil {
			return err
		}

		keys, err := curr.Keys(ctx)

		if err != nil {
			return err
		}

		keySlice, err := ToRuntimeStringSlice(ctx, keys)

		if err != nil {
			return err
		}

		err = el.RemoveStyle(ctx, keySlice...)

		if err != nil {
			return err
		}

		return obj.ForEach(ctx, func(ctx context.Context, value, key runtime.Value) (runtime.Boolean, error) {
			err = el.SetStyle(ctx, runtime.NewString(key.String()), runtime.NewString(value.String()))

			if err != nil {
				return false, err
			}

			return true, nil
		})
	case "value":
		err := el.SetValue(ctx, value)

		if err != nil {
			return err
		}

		return nil
	}

	return ErrInvalidPath
}

func SetInNode(_ context.Context, _ runtime.Value, _ drivers.HTMLNode, _ runtime.Value) error {
	return ErrReadOnly
}
