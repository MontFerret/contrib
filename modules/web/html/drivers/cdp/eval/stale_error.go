package eval

import (
	"errors"
	"strings"

	"github.com/mafredri/cdp/rpcc"
)

var staleResponseMessages = []string{
	"cannot find context with specified id",
	"execution context was destroyed",
	"execution context with given id not found",
	"argument should belong to the same javascript world",
	"could not find object with given id",
	"could not find node with given id",
	"no node with given id found",
	"node with given id does not belong to the document",
}

// IsStaleError reports whether Chrome rejected a request because its document
// execution state was invalidated before the operation could start.
func IsStaleError(err error) bool {
	var responseErr *rpcc.ResponseError
	if !errors.As(err, &responseErr) || responseErr.Code != -32000 {
		return false
	}

	message := strings.ToLower(responseErr.Message + " " + responseErr.Data)
	for _, known := range staleResponseMessages {
		if strings.Contains(message, known) {
			return true
		}
	}

	return false
}
