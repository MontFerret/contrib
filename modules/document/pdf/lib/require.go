package lib

import (
	"fmt"

	"github.com/MontFerret/contrib/modules/document/pdf/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func requireDocument(value runtime.Value, operation string) (*core.Document, error) {
	document, ok := value.(*core.Document)
	if !ok {
		return nil, fmt.Errorf("DOCUMENT::PDF %s failed: expected PDF document handle", operation)
	}

	return document, nil
}

func requirePage(value runtime.Value, operation string) (*core.Page, error) {
	page, ok := value.(*core.Page)
	if !ok {
		return nil, fmt.Errorf("DOCUMENT::PDF %s failed: expected PDF page handle", operation)
	}

	return page, nil
}

func requireString(value runtime.Value, operation, name string) (string, error) {
	str, ok := value.(runtime.String)
	if !ok {
		return "", fmt.Errorf("DOCUMENT::PDF %s failed: %s must be a string", operation, name)
	}

	return str.String(), nil
}

func requireInt(value runtime.Value, operation, name string) (int, error) {
	integer, ok := value.(runtime.Int)
	if !ok {
		return 0, fmt.Errorf("DOCUMENT::PDF %s failed: %s must be an integer", operation, name)
	}

	return int(integer), nil
}
