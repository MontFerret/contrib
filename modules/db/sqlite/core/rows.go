package core

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func scanRows(ctx context.Context, rows *sql.Rows) (runtime.List, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	out := runtime.NewArray(0)

	for rows.Next() {
		values := make([]any, len(columns))
		dest := make([]any, len(columns))

		for idx := range values {
			dest[idx] = &values[idx]
		}

		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}

		obj := runtime.NewObjectOf(len(columns))
		for idx, column := range columns {
			value, err := sqlValueToRuntime(values[idx])

			if err != nil {
				return nil, err
			}

			if err := obj.Set(ctx, runtime.NewString(column), value); err != nil {
				return nil, err
			}
		}

		if err := out.Append(ctx, obj); err != nil {
			return nil, err
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func sqlValueToRuntime(value any) (runtime.Value, error) {
	switch val := value.(type) {
	case nil:
		return runtime.None, nil
	case int:
		return runtime.NewInt(val), nil
	case int64:
		return runtime.NewInt64(val), nil
	case int32:
		return runtime.NewInt64(int64(val)), nil
	case int16:
		return runtime.NewInt64(int64(val)), nil
	case int8:
		return runtime.NewInt64(int64(val)), nil
	case uint:
		return runtime.NewInt64(int64(val)), nil
	case uint64:
		return runtime.NewInt64(int64(val)), nil
	case uint32:
		return runtime.NewInt64(int64(val)), nil
	case uint16:
		return runtime.NewInt64(int64(val)), nil
	case uint8:
		return runtime.NewInt64(int64(val)), nil
	case float64:
		return runtime.NewFloat(val), nil
	case float32:
		return runtime.NewFloat(float64(val)), nil
	case string:
		return runtime.NewString(val), nil
	case bool:
		return runtime.NewBoolean(val), nil
	case []byte:
		out := make([]byte, len(val))
		copy(out, val)

		return runtime.NewBinary(out), nil
	default:
		return runtime.None, fmt.Errorf("unsupported SQLite value type %T", value)
	}
}
