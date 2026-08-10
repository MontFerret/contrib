package network

import (
	"errors"
	"strings"

	"github.com/mafredri/cdp/rpcc"
)

var unavailableResponseBodyMessages = []string{
	"no resource with given identifier found",
	"no data found for resource with given identifier",
	"request content was evicted from inspector cache",
	"can only get response body on requests captured after network.enable",
}

func isUnavailableResponseBodyError(err error) bool {
	var responseErr *rpcc.ResponseError
	if !errors.As(err, &responseErr) || responseErr.Code != -32000 {
		return false
	}

	message := strings.ToLower(responseErr.Message + " " + responseErr.Data)

	for _, known := range unavailableResponseBodyMessages {
		if strings.Contains(message, known) {
			return true
		}
	}

	return false
}
