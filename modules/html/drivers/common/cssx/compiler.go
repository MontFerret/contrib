package cssx

import (
	"errors"
	"fmt"
	"strings"

	"github.com/MontFerret/cssx"
)

func Compile(input string) (cssx.Pipeline, error) {
	return cssx.Compile(input)
}

func ResolveSelector(selector string) (Expression, error) {
	value := strings.TrimSpace(selector)
	if value == "" {
		return "", errors.New("selector is empty")
	}

	if !strings.HasPrefix(value, ":") {
		value = ":" + value
	}

	resolved, ok := selectorLookup[value]
	if !ok {
		return "", fmt.Errorf("unknown selector %q", value)
	}

	return resolved, nil
}
