package lib

import (
	"context"
	"fmt"

	"github.com/MontFerret/contrib/modules/document/xlsx/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Create creates a new in-memory XLSX workbook.
func Create(context.Context) (runtime.Value, error) {
	return core.Create(), nil
}

// Open opens an existing XLSX workbook.
func Open(ctx context.Context, pathValue runtime.Value) (runtime.Value, error) {
	path, ok := pathValue.(runtime.String)
	if !ok {
		return runtime.None, fmt.Errorf("DOCUMENT::XLSX OPEN failed: path must be a string")
	}

	return core.Open(ctx, path.String())
}

// Sheets returns worksheet names in workbook order.
func Sheets(ctx context.Context, workbookValue runtime.Value) (runtime.Value, error) {
	workbook, err := requireWorkbook(workbookValue, "SHEETS")
	if err != nil {
		return runtime.None, err
	}

	names, err := workbook.Sheets()
	if err != nil {
		return runtime.None, err
	}

	out := runtime.NewArray(len(names))
	for _, name := range names {
		if err := out.Append(ctx, runtime.NewString(name)); err != nil {
			return runtime.None, err
		}
	}

	return out, nil
}

// Sheet returns a worksheet handle by name.
func Sheet(_ context.Context, workbookValue, nameValue runtime.Value) (runtime.Value, error) {
	workbook, err := requireWorkbook(workbookValue, "SHEET")
	if err != nil {
		return runtime.None, err
	}

	name, err := requireString(nameValue, "SHEET", "name")
	if err != nil {
		return runtime.None, err
	}

	return workbook.Sheet(name)
}

// AddSheet creates a worksheet and returns its handle.
func AddSheet(_ context.Context, workbookValue, nameValue runtime.Value) (runtime.Value, error) {
	workbook, err := requireWorkbook(workbookValue, "ADD_SHEET")
	if err != nil {
		return runtime.None, err
	}

	name, err := requireString(nameValue, "ADD_SHEET", "name")
	if err != nil {
		return runtime.None, err
	}

	return workbook.AddSheet(name)
}

// DeleteSheet deletes a worksheet.
func DeleteSheet(_ context.Context, workbookValue, nameValue runtime.Value) (runtime.Value, error) {
	workbook, err := requireWorkbook(workbookValue, "DELETE_SHEET")
	if err != nil {
		return runtime.None, err
	}

	name, err := requireString(nameValue, "DELETE_SHEET", "name")
	if err != nil {
		return runtime.None, err
	}

	if err := workbook.DeleteSheet(name); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}

// Save saves a workbook to its current path.
func Save(ctx context.Context, workbookValue runtime.Value) (runtime.Value, error) {
	workbook, err := requireWorkbook(workbookValue, "SAVE")
	if err != nil {
		return runtime.None, err
	}
	if err := workbook.Save(ctx); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}

// SaveAs saves a workbook to a supplied path.
func SaveAs(ctx context.Context, workbookValue, pathValue runtime.Value) (runtime.Value, error) {
	workbook, err := requireWorkbook(workbookValue, "SAVE_AS")
	if err != nil {
		return runtime.None, err
	}

	path, err := requireString(pathValue, "SAVE_AS", "path")
	if err != nil {
		return runtime.None, err
	}
	if err := workbook.SaveAs(ctx, path); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}

// Close releases workbook resources.
func Close(_ context.Context, workbookValue runtime.Value) (runtime.Value, error) {
	workbook, err := requireWorkbook(workbookValue, "CLOSE")
	if err != nil {
		return runtime.None, err
	}
	if err := workbook.Close(); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}
