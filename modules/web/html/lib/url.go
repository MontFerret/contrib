package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// URL returns the current document URL.
// @param {HTMLPage | HTMLDocument} root - HTML page or document.
// @return {String} - Current document URL.
func URL(ctx context.Context, root runtime.Value) (runtime.Value, error) {
	target, err := drivers.ToDocumentURLTarget(root)
	if err != nil {
		return runtime.None, err
	}

	return target.GetCurrentURL(ctx)
}

// BaseURL returns the effective document base URL.
// @param {HTMLPage | HTMLDocument} root - HTML page or document.
// @return {String} - Effective document base URL.
func BaseURL(ctx context.Context, root runtime.Value) (runtime.Value, error) {
	target, err := drivers.ToDocumentURLTarget(root)
	if err != nil {
		return runtime.None, err
	}

	return target.GetBaseURL(ctx)
}

// ResolveURL resolves a URL against the effective document base URL.
// @param {HTMLPage | HTMLDocument} root - HTML page or document.
// @param {String} url - URL to resolve.
// @return {String} - Resolved URL.
func ResolveURL(ctx context.Context, root, url runtime.Value) (runtime.Value, error) {
	target, err := drivers.ToDocumentURLTarget(root)
	if err != nil {
		return runtime.None, err
	}

	value, err := runtime.CastString(url)
	if err != nil {
		return runtime.None, err
	}

	return target.ResolveURL(ctx, value)
}
