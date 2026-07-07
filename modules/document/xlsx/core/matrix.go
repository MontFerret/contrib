package core

import (
	"context"
	"fmt"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func decodeMatrix(ctx context.Context, value runtime.Value) ([][]runtime.Value, error) {
	rowsList, ok := value.(runtime.List)
	if !ok {
		return nil, fmt.Errorf("rows must be an array of row arrays")
	}

	rows := make([][]runtime.Value, 0)
	expectedColumns := -1
	rowIndex := 0

	err := runtime.ForEach(ctx, rowsList, func(ctx context.Context, value, _ runtime.Value) (runtime.Boolean, error) {
		rowIndex++

		rowList, ok := value.(runtime.List)
		if !ok {
			return runtime.False, fmt.Errorf("row %d must be an array", rowIndex)
		}

		row := make([]runtime.Value, 0)
		if err := runtime.ForEach(ctx, rowList, func(_ context.Context, value, _ runtime.Value) (runtime.Boolean, error) {
			if value == nil {
				value = runtime.None
			}

			row = append(row, value)

			return runtime.True, nil
		}); err != nil {
			return runtime.False, err
		}

		if len(row) == 0 {
			return runtime.False, fmt.Errorf("row %d has 0 cells; expected at least 1", rowIndex)
		}

		if expectedColumns < 0 {
			expectedColumns = len(row)
		} else if len(row) != expectedColumns {
			return runtime.False, fmt.Errorf("row %d has %d cells; expected %d", rowIndex, len(row), expectedColumns)
		}

		rows = append(rows, row)

		return runtime.True, nil
	})

	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("rows must contain at least one row")
	}

	return rows, nil
}
