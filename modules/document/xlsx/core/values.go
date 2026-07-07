package core

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func readCellValue(file *excelize.File, sheet, cell string) (runtime.Value, error) {
	cellType, err := file.GetCellType(sheet, cell)
	if err != nil {
		return runtime.None, err
	}

	if cellType == excelize.CellTypeUnset {
		value, err := file.GetCellValue(sheet, cell, excelize.Options{RawCellValue: true})

		if err != nil {
			return runtime.None, err
		}

		if value == "" {
			return runtime.None, nil
		}

		if date, err := styledDateCellValue(file, sheet, cell, value); err == nil {
			return date, nil
		}

		return inferCachedFormulaValue(value), nil
	}

	if cellType == excelize.CellTypeError {
		return runtime.None, nil
	}

	switch cellType {
	case excelize.CellTypeBool:
		value, err := file.GetCellValue(sheet, cell, excelize.Options{RawCellValue: true})

		if err != nil {
			return runtime.None, err
		}

		return boolCellValue(value)
	case excelize.CellTypeNumber:
		value, err := file.GetCellValue(sheet, cell, excelize.Options{RawCellValue: true})

		if err != nil {
			return runtime.None, err
		}

		return numericCellValue(value)
	case excelize.CellTypeDate:
		value, err := dateCellValue(file, sheet, cell)

		if err == nil {
			return value, nil
		}

		text, textErr := file.GetCellValue(sheet, cell)

		if textErr != nil {
			return runtime.None, textErr
		}

		return runtime.NewString(text), nil
	case excelize.CellTypeFormula:
		text, err := file.GetCellValue(sheet, cell)

		if err != nil {
			return runtime.None, err
		}

		return inferCachedFormulaValue(text), nil
	default:
		text, err := file.GetCellValue(sheet, cell)

		if err != nil {
			return runtime.None, err
		}

		return runtime.NewString(text), nil
	}
}

func writeCellValue(file *excelize.File, sheet, cell string, value runtime.Value) error {
	if value == nil || value == runtime.None {
		return file.SetCellValue(sheet, cell, nil)
	}

	switch val := value.(type) {
	case runtime.String:
		return file.SetCellStr(sheet, cell, val.String())
	case runtime.Boolean:
		return file.SetCellBool(sheet, cell, bool(val))
	case runtime.Int:
		return file.SetCellInt(sheet, cell, int64(val))
	case runtime.Float:
		return file.SetCellFloat(sheet, cell, float64(val), -1, 64)
	case runtime.DateTime:
		return file.SetCellValue(sheet, cell, val.Time)
	default:
		return fmt.Errorf("unsupported XLSX cell value type %s", runtime.TypeName(runtime.TypeOf(value)))
	}
}

func boolCellValue(value string) (runtime.Value, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true":
		return runtime.True, nil
	case "0", "false":
		return runtime.False, nil
	default:
		return runtime.None, fmt.Errorf("invalid boolean cell value %q", value)
	}
}

func numericCellValue(value string) (runtime.Value, error) {
	number, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return runtime.None, err
	}

	if math.IsNaN(number) || math.IsInf(number, 0) {
		return runtime.NewFloat(number), nil
	}

	if math.Trunc(number) == number && number >= math.MinInt64 && number <= math.MaxInt64 {
		return runtime.NewInt64(int64(number)), nil
	}

	return runtime.NewFloat(number), nil
}

func dateCellValue(file *excelize.File, sheet, cell string) (runtime.Value, error) {
	raw, err := file.GetCellValue(sheet, cell, excelize.Options{RawCellValue: true})

	if err != nil {
		return runtime.None, err
	}

	return dateCellValueFromSerial(file, raw)
}

func styledDateCellValue(file *excelize.File, sheet, cell, raw string) (runtime.Value, error) {
	styleID, err := file.GetCellStyle(sheet, cell)
	if err != nil {
		return runtime.None, err
	}

	style, err := file.GetStyle(styleID)

	if err != nil {
		return runtime.None, err
	}

	if style == nil || !isDateStyle(style) {
		return runtime.None, fmt.Errorf("cell is not styled as date")
	}

	return dateCellValueFromSerial(file, raw)
}

func dateCellValueFromSerial(file *excelize.File, raw string) (runtime.Value, error) {
	serial, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return runtime.None, err
	}

	props, err := file.GetWorkbookProps()
	if err != nil {
		return runtime.None, err
	}

	use1904 := false
	if props.Date1904 != nil {
		use1904 = *props.Date1904
	}

	value, err := excelize.ExcelDateToTime(serial, use1904)
	if err != nil {
		return runtime.None, err
	}

	return runtime.NewDateTime(value), nil
}

func isDateStyle(style *excelize.Style) bool {
	switch style.NumFmt {
	case 14, 15, 16, 17, 18, 19, 20, 21, 22, 45, 46, 47:
		return true
	}

	if style.CustomNumFmt == nil {
		return false
	}

	format := strings.ToLower(*style.CustomNumFmt)

	return strings.Contains(format, "yy") ||
		strings.Contains(format, "dd") ||
		strings.Contains(format, "hh") ||
		strings.Contains(format, "ss")
}

func inferCachedFormulaValue(text string) runtime.Value {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return runtime.None
	}

	if value, err := boolCellValue(trimmed); err == nil {
		return value
	}

	if value, err := numericCellValue(trimmed); err == nil {
		return value
	}

	if value, err := time.Parse(runtime.DefaultTimeLayout, trimmed); err == nil {
		return runtime.NewDateTime(value)
	}

	return runtime.NewString(text)
}

func runtimeRowsToArray(ctx context.Context, rows [][]runtime.Value) (*runtime.Array, error) {
	out := runtime.NewArray(len(rows))

	for _, row := range rows {
		rowArray := runtime.NewArray(len(row))

		for _, value := range row {
			if err := rowArray.Append(ctx, value); err != nil {
				return nil, err
			}
		}

		if err := out.Append(ctx, rowArray); err != nil {
			return nil, err
		}
	}

	return out, nil
}
