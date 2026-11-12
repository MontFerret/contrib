// Package origin matches HTTP(S) request origins for the HTML drivers.
package origin

import (
	"fmt"
	"net/url"
	"strings"
)

// Allowlist is an immutable set of normalized HTTP(S) origins.
type Allowlist struct {
	origins map[origin]struct{}
}

// NewAllowlist returns nil for an omitted policy and an empty set for deny-all.
func NewAllowlist(values []string) (*Allowlist, error) {
	if values == nil {
		return nil, nil
	}

	list := &Allowlist{origins: make(map[origin]struct{}, len(values))}
	for i, value := range values {
		parsed, err := url.Parse(value)
		if err != nil {
			// Parser errors can contain credentials from the configuration.
			return nil, fmt.Errorf("allowed origin %d: expected an absolute HTTP(S) URL", i+1)
		}

		if parsed.User != nil || (parsed.Path != "" && parsed.Path != "/") ||
			parsed.RawPath != "" || parsed.ForceQuery || parsed.RawQuery != "" || strings.Contains(value, "#") {
			return nil, fmt.Errorf("allowed origin %d: expected an origin without credentials, path, query, or fragment", i+1)
		}

		key, err := normalize(parsed)
		if err != nil {
			return nil, fmt.Errorf("allowed origin %d: %w", i+1, err)
		}

		list.origins[key] = struct{}{}
	}

	return list, nil
}

// Match returns the normalized hostname if the request URL belongs to a listed origin.
func (list *Allowlist) Match(u *url.URL) (string, bool) {
	key, err := normalize(u)
	if err != nil {
		return "", false
	}

	_, allowed := list.origins[key]

	return key.host, allowed
}
