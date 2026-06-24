package core

import (
	"fmt"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const dialectHTTP = "http"

func validateQueryDialect(kind runtime.String) error {
	dialect := strings.TrimSpace(kind.String())
	if dialect == "" || strings.EqualFold(dialect, dialectHTTP) {
		return nil
	}

	return fmt.Errorf("unsupported dialect %q; expected %q", dialect, dialectHTTP)
}
