package core

import (
	"context"
	"fmt"

	commonresource "github.com/MontFerret/contrib/pkg/common/resource"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Worksheet is an opaque XLSX worksheet handle exposed to Ferret.
type Worksheet struct {
	workbook   *Workbook
	name       string
	generation uint64
	id         uint64
}

// NewWorksheet creates a tracked XLSX worksheet handle.
func NewWorksheet(workbook *Workbook, name string, generation uint64) *Worksheet {
	return &Worksheet{
		workbook:   workbook,
		name:       name,
		generation: generation,
		id:         newResourceID(),
	}
}

func (s *Worksheet) Get(cell string) (runtime.Value, error) {
	ref, err := parseCellRef(cell)
	if err != nil {
		return runtime.None, OperationError("GET", err)
	}

	s.workbook.mu.Lock()
	defer s.workbook.mu.Unlock()

	if err := s.workbook.validateWorksheet(s.name, s.generation); err != nil {
		return runtime.None, OperationError("GET", err)
	}

	value, err := readCellValue(s.workbook.file, s.name, ref.Name)
	if err != nil {
		return runtime.None, OperationErrorf("GET", "cell %q: %w", ref.Name, err)
	}

	return value, nil
}

func (s *Worksheet) Set(cell string, value runtime.Value) error {
	ref, err := parseCellRef(cell)
	if err != nil {
		return OperationError("SET", err)
	}

	s.workbook.mu.Lock()
	defer s.workbook.mu.Unlock()

	if err := s.workbook.validateWorksheet(s.name, s.generation); err != nil {
		return OperationError("SET", err)
	}
	if err := writeCellValue(s.workbook.file, s.name, ref.Name, value); err != nil {
		return OperationErrorf("SET", "cell %q: %w", ref.Name, err)
	}

	return nil
}

func (s *Worksheet) Range(ctx context.Context, ref string) (runtime.List, error) {
	rows, err := s.rangeRows(ref)
	if err != nil {
		return nil, err
	}

	return runtimeRowsToArray(ctx, rows)
}

func (s *Worksheet) WriteRange(ctx context.Context, ref string, value runtime.Value) error {
	rangeRef, err := parseRangeRef(ref)
	if err != nil {
		return OperationError("WRITE_RANGE", err)
	}

	rows, err := decodeMatrix(ctx, value)
	if err != nil {
		return OperationError("WRITE_RANGE", err)
	}
	if !rangeRef.Single {
		expectedRows := rangeRef.End.Row - rangeRef.Start.Row + 1
		expectedCols := rangeRef.End.Col - rangeRef.Start.Col + 1
		if len(rows) != expectedRows || len(rows[0]) != expectedCols {
			return OperationErrorf(
				"WRITE_RANGE",
				"range %q expects %d rows and %d columns; got %d rows and %d columns",
				ref,
				expectedRows,
				expectedCols,
				len(rows),
				len(rows[0]),
			)
		}
	}

	s.workbook.mu.Lock()
	defer s.workbook.mu.Unlock()

	if err := s.workbook.validateWorksheet(s.name, s.generation); err != nil {
		return OperationError("WRITE_RANGE", err)
	}

	return s.writeRows(rangeRef.Start.Col, rangeRef.Start.Row, rows)
}

func (s *Worksheet) Append(ctx context.Context, value runtime.Value) error {
	rows, err := decodeMatrix(ctx, value)
	if err != nil {
		return OperationError("APPEND", err)
	}

	s.workbook.mu.Lock()
	defer s.workbook.mu.Unlock()

	if err := s.workbook.validateWorksheet(s.name, s.generation); err != nil {
		return OperationError("APPEND", err)
	}

	startRow, err := s.lastPopulatedRow()
	if err != nil {
		return OperationError("APPEND", err)
	}

	return s.writeRows(1, startRow+1, rows)
}

func (s *Worksheet) Query(ctx context.Context, q runtime.Query) (runtime.List, error) {
	if q.Kind != runtime.EmptyString {
		return nil, OperationErrorf("QUERY", "unsupported XLSX query dialect %q", q.Kind.String())
	}

	opts, err := decodeQueryOptions(ctx, q.Params)
	if err != nil {
		return nil, OperationError("QUERY", err)
	}

	rows, err := s.rangeRows(q.Expression.String())
	if err != nil {
		return nil, err
	}

	out, err := applyQueryOptions(ctx, rows, opts)
	if err != nil {
		return nil, OperationError("QUERY", err)
	}

	return out, nil
}

func (s *Worksheet) QueryOne(ctx context.Context, q runtime.Query) (runtime.Value, error) {
	return runtime.DefaultQueryOne(ctx, q, s.Query)
}

func (s *Worksheet) QueryCount(ctx context.Context, q runtime.Query) (runtime.Int, error) {
	return runtime.DefaultQueryCount(ctx, q, s.Query)
}

func (s *Worksheet) QueryExists(ctx context.Context, q runtime.Query) (runtime.Boolean, error) {
	return runtime.DefaultQueryExists(ctx, q, s.Query)
}

func (s *Worksheet) ResourceID() uint64 {
	return s.id
}

func (s *Worksheet) String() string {
	return fmt.Sprintf("XLSXSheet(%q)", s.name)
}

func (s *Worksheet) Hash() uint64 {
	return commonresource.Hash("document.xlsx.worksheet", s.id)
}

func (s *Worksheet) Copy() runtime.Value {
	return s
}

func (s *Worksheet) MarshalJSON() ([]byte, error) {
	return commonresource.MarshalStringJSON(s.String())
}

func (s *Worksheet) rangeRows(ref string) ([][]runtime.Value, error) {
	rangeRef, err := parseRangeRef(ref)
	if err != nil {
		return nil, OperationError("RANGE", err)
	}

	s.workbook.mu.Lock()
	defer s.workbook.mu.Unlock()

	if err := s.workbook.validateWorksheet(s.name, s.generation); err != nil {
		return nil, OperationError("RANGE", err)
	}

	rows := make([][]runtime.Value, 0, rangeRef.End.Row-rangeRef.Start.Row+1)
	for row := rangeRef.Start.Row; row <= rangeRef.End.Row; row++ {
		outRow := make([]runtime.Value, 0, rangeRef.End.Col-rangeRef.Start.Col+1)

		for col := rangeRef.Start.Col; col <= rangeRef.End.Col; col++ {
			name, err := cellName(col, row)
			if err != nil {
				return nil, OperationError("RANGE", err)
			}

			value, err := readCellValue(s.workbook.file, s.name, name)
			if err != nil {
				return nil, OperationErrorf("RANGE", "cell %q: %w", name, err)
			}

			outRow = append(outRow, value)
		}

		rows = append(rows, outRow)
	}

	return rows, nil
}

func (s *Worksheet) writeRows(startCol, startRow int, rows [][]runtime.Value) error {
	for rowIdx, row := range rows {
		for colIdx, value := range row {
			name, err := cellName(startCol+colIdx, startRow+rowIdx)

			if err != nil {
				return OperationError("WRITE_RANGE", err)
			}

			if err := writeCellValue(s.workbook.file, s.name, name, value); err != nil {
				return OperationErrorf(
					"WRITE_RANGE",
					"row %d column %d cell %q: %w",
					rowIdx+1,
					colIdx+1,
					name,
					err,
				)
			}
		}
	}

	return nil
}

func (s *Worksheet) lastPopulatedRow() (int, error) {
	rows, err := s.workbook.file.GetRows(s.name)
	if err != nil {
		return 0, err
	}

	for idx := len(rows) - 1; idx >= 0; idx-- {
		for _, value := range rows[idx] {
			if value != "" {
				return idx + 1, nil
			}
		}
	}

	return 0, nil
}
