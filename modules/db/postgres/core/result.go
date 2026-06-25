package core

import (
	"database/sql"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func execResult(result sql.Result) runtime.Value {
	rowsAffected := runtime.Value(runtime.None)
	if value, err := result.RowsAffected(); err == nil {
		rowsAffected = runtime.NewInt64(value)
	}

	return runtime.NewObjectWith(map[string]runtime.Value{
		"rowsAffected": rowsAffected,
		"lastInsertId": runtime.None,
	})
}
