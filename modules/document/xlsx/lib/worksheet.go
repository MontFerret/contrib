package lib

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Get reads a single worksheet cell.
func Get(_ context.Context, sheetValue, cellValue runtime.Value) (runtime.Value, error) {
	sheet, err := requireWorksheet(sheetValue, "GET")
	if err != nil {
		return runtime.None, err
	}

	cell, err := requireString(cellValue, "GET", "cell")
	if err != nil {
		return runtime.None, err
	}

	return sheet.Get(cell)
}

// Set writes a single worksheet cell.
func Set(_ context.Context, sheetValue, cellValue, value runtime.Value) (runtime.Value, error) {
	sheet, err := requireWorksheet(sheetValue, "SET")
	if err != nil {
		return runtime.None, err
	}

	cell, err := requireString(cellValue, "SET", "cell")
	if err != nil {
		return runtime.None, err
	}
	if err := sheet.Set(cell, value); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}

// Range reads a worksheet range as row arrays.
func Range(ctx context.Context, sheetValue, rangeValue runtime.Value) (runtime.Value, error) {
	sheet, err := requireWorksheet(sheetValue, "RANGE")
	if err != nil {
		return runtime.None, err
	}

	ref, err := requireString(rangeValue, "RANGE", "range")
	if err != nil {
		return runtime.None, err
	}

	return sheet.Range(ctx, ref)
}

// WriteRange writes a rectangular matrix into a worksheet range.
func WriteRange(ctx context.Context, sheetValue, rangeValue, rowsValue runtime.Value) (runtime.Value, error) {
	sheet, err := requireWorksheet(sheetValue, "WRITE_RANGE")
	if err != nil {
		return runtime.None, err
	}

	ref, err := requireString(rangeValue, "WRITE_RANGE", "range")
	if err != nil {
		return runtime.None, err
	}
	if err := sheet.WriteRange(ctx, ref, rowsValue); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}

// Append appends rows after the last populated worksheet row.
func Append(ctx context.Context, sheetValue, rowsValue runtime.Value) (runtime.Value, error) {
	sheet, err := requireWorksheet(sheetValue, "APPEND")
	if err != nil {
		return runtime.None, err
	}
	if err := sheet.Append(ctx, rowsValue); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}
