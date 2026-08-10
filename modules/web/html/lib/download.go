package lib

import (
	"context"
	"io"
	"net/http"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Download downloads a resource from a URL.
//
// @param url {String} Resource URL.
// @return {Binary} Downloaded bytes.
func Download(_ context.Context, url runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateType(url, runtime.TypeString)

	if err != nil {
		return runtime.None, err
	}

	resp, err := http.Get(url.String())

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
