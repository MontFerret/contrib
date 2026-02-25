package html

import (
	"context"
	"io"
	"net/http"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// DOWNLOAD downloads a resource from the given GetURL.
// @param {String} url - URL to download.
// @return {Binary} - A base64 encoded string in binary format.
func Download(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 1, 1)

	if err != nil {
		return runtime.None, err
	}

	arg1 := args[0]
	err = runtime.ValidateType(arg1, runtime.TypeString)

	if err != nil {
		return runtime.None, err
	}

	resp, err := http.Get(arg1.String())

	if err != nil {
		return runtime.None, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return runtime.None, err
	}

	return runtime.NewBinary(data), nil
}
