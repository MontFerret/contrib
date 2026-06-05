package core

import (
	"errors"
	"net/url"
	"strings"
)

const fileDBDisabledMessage = "file-backed SQLite databases are disabled; use memory: true or uri with mode=memory"

// OpenPolicy controls which SQLite database sources may be opened.
type OpenPolicy struct {
	memoryOnly bool
}

// DefaultOpenPolicy allows all supported SQLite database sources.
func DefaultOpenPolicy() OpenPolicy {
	return OpenPolicy{}
}

// MemoryOnlyOpenPolicy allows private memory databases and explicit
// mode=memory SQLite file URIs.
func MemoryOnlyOpenPolicy() OpenPolicy {
	return OpenPolicy{
		memoryOnly: true,
	}
}

func (p OpenPolicy) validate(options OpenOptions) error {
	if !p.memoryOnly {
		return nil
	}

	if options.pathProvided() {
		return errors.New(fileDBDisabledMessage)
	}
	if options.uriProvided() && !isMemoryURI(*options.URI) {
		return errors.New(fileDBDisabledMessage)
	}

	return nil
}

func isMemoryURI(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return false
	}

	return strings.EqualFold(parsed.Scheme, "file") && strings.EqualFold(parsed.Query().Get("mode"), "memory")
}
