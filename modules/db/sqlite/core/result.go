package core

import (
	"database/sql"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func execResult(sqlText string, result sql.Result) runtime.Value {
	rowsAffected := runtime.Value(runtime.None)
	if value, err := result.RowsAffected(); err == nil {
		rowsAffected = runtime.NewInt64(value)
	}

	lastInsertID := runtime.Value(runtime.None)
	if isInsertStatement(sqlText) {
		if value, err := result.LastInsertId(); err == nil {
			lastInsertID = runtime.NewInt64(value)
		}
	}

	return runtime.NewObjectWith(map[string]runtime.Value{
		"rowsAffected": rowsAffected,
		"lastInsertId": lastInsertID,
	})
}

func isInsertStatement(sqlText string) bool {
	fields := strings.Fields(strings.TrimSpace(sqlText))
	if len(fields) == 0 {
		return false
	}

	switch strings.ToUpper(fields[0]) {
	case "INSERT", "REPLACE":
		return true
	default:
		return false
	}
}
