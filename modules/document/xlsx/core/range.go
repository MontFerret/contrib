package core

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

type (
	CellRef struct {
		Name string
		Col  int
		Row  int
	}

	RangeRef struct {
		Start  CellRef
		End    CellRef
		Single bool
	}
)

func parseCellRef(input string) (CellRef, error) {
	cell := strings.TrimSpace(input)
	if cell == "" {
		return CellRef{}, fmt.Errorf("cell reference is empty")
	}

	if strings.ContainsAny(cell, "$!: ") {
		return CellRef{}, fmt.Errorf("invalid XLSX cell reference %q", input)
	}

	col, row, err := excelize.CellNameToCoordinates(cell)
	if err != nil || col < 1 || row < 1 {
		return CellRef{}, fmt.Errorf("invalid XLSX cell reference %q", input)
	}

	name, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return CellRef{}, fmt.Errorf("invalid XLSX cell reference %q", input)
	}

	return CellRef{Name: name, Col: col, Row: row}, nil
}

func parseRangeRef(input string) (RangeRef, error) {
	text := strings.TrimSpace(input)

	if text == "" {
		return RangeRef{}, fmt.Errorf("range reference is empty")
	}

	if strings.Count(text, ":") > 1 {
		return RangeRef{}, fmt.Errorf("invalid XLSX range %q", input)
	}

	parts := strings.Split(text, ":")
	start, err := parseCellRef(parts[0])
	if err != nil {
		return RangeRef{}, fmt.Errorf("invalid XLSX range %q: %w", input, err)
	}

	if len(parts) == 1 {
		return RangeRef{Start: start, End: start, Single: true}, nil
	}

	end, err := parseCellRef(parts[1])
	if err != nil {
		return RangeRef{}, fmt.Errorf("invalid XLSX range %q: %w", input, err)
	}

	if start.Col > end.Col || start.Row > end.Row {
		return RangeRef{}, fmt.Errorf("invalid XLSX range %q: start must be above and left of end", input)
	}

	return RangeRef{Start: start, End: end}, nil
}

func cellName(col, row int) (string, error) {
	name, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return "", err
	}

	return name, nil
}
