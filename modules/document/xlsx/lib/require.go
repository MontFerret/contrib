package lib

import (
	"fmt"

	"github.com/MontFerret/contrib/modules/document/xlsx/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func requireWorkbook(value runtime.Value, operation string) (*core.Workbook, error) {
	workbook, ok := value.(*core.Workbook)
	if !ok {
		return nil, fmt.Errorf("DOCUMENT::XLSX %s failed: expected XLSX workbook handle", operation)
	}

	return workbook, nil
}

func requireWorksheet(value runtime.Value, operation string) (*core.Worksheet, error) {
	sheet, ok := value.(*core.Worksheet)
	if !ok {
		return nil, fmt.Errorf("DOCUMENT::XLSX %s failed: expected XLSX worksheet handle", operation)
	}

	return sheet, nil
}

func requireString(value runtime.Value, operation, name string) (string, error) {
	str, ok := value.(runtime.String)
	if !ok {
		return "", fmt.Errorf("DOCUMENT::XLSX %s failed: %s must be a string", operation, name)
	}

	return str.String(), nil
}
