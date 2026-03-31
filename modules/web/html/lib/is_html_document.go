package lib

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// IsHTMLDocument checks whether a value is an HTMLDocument value.
// @param {Any} value - Input value of arbitrary type.
// @return {Boolean} - Returns true if value is HTMLDocument, otherwise false.
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
