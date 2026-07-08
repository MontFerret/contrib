package lib

import (
	"fmt"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func requireString(value runtime.Value, operation, name string) (string, error) {
	str, ok := value.(runtime.String)
	if !ok {
		return "", fmt.Errorf("DOCUMENT::PDF %s failed: %s must be a string", operation, name)
	}

	return str.String(), nil
}
