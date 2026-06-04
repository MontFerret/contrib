package core

import (
	"fmt"
	"strings"
)

const dialectSQL = "sql"
const dialectSQLExec = "sql_exec"

type queryDialect int

const (
	queryDialectRows queryDialect = iota
	queryDialectExec
)

func validateDialect(dialect string) error {
	if strings.EqualFold(dialect, dialectSQL) {
		return nil
	}

	return fmt.Errorf("unsupported dialect %q; expected %q", dialect, dialectSQL)
}

func parseQueryDialect(dialect string) (queryDialect, error) {
	if strings.EqualFold(dialect, dialectSQL) {
		return queryDialectRows, nil
	}
	if strings.EqualFold(dialect, dialectSQLExec) {
		return queryDialectExec, nil
	}

	return queryDialectRows, fmt.Errorf(
		"unsupported dialect %q; expected %q or %q",
		dialect,
		dialectSQL,
		dialectSQLExec,
	)
}
