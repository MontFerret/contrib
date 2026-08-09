package core

import (
	"fmt"
	"strings"
)

type queryDialect int

const (
	queryDialectRead queryDialect = iota
	queryDialectExec
)

const (
	dialectRedis     = "redis"
	dialectRedisExec = "redis_exec"
)

func parseQueryDialect(dialect string) (queryDialect, error) {
	if dialect == "" {
		// Fall back to the default dialect if none is specified
		return queryDialectRead, nil
	}

	if strings.EqualFold(dialect, dialectRedis) {
		return queryDialectRead, nil
	}
	if strings.EqualFold(dialect, dialectRedisExec) {
		return queryDialectExec, nil
	}

	return queryDialectRead, fmt.Errorf(
		"unsupported dialect %q; expected %q or %q",
		dialect,
		dialectRedis,
		dialectRedisExec,
	)
}
