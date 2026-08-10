package lib

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// IsHTMLDocument reports whether a value is an HTML document.
//
// @param value {Any} Value to inspect.
// @return {Boolean} Whether the value is an HTMLDocument.
func IsHTMLDocument(_ context.Context, arg runtime.Value) (runtime.Value, error) {
	//err := runtime.ValidateArgs(args, 1, 1)
	//
	//if err != nil {
	//	return runtime.None, err
	//}
	//
	//return isTypeof(args[0], drivers.HTMLDocumentType), nil

	panic("implement me")
}
