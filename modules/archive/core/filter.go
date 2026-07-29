package core

import (
	"fmt"
	"path"
)

type entryFilter struct {
	include []string
	exclude []string
}

func newEntryFilter(include, exclude []string) (*entryFilter, error) {
	for _, pattern := range append(append([]string(nil), include...), exclude...) {
		if _, err := path.Match(pattern, ""); err != nil {
			return nil, fmt.Errorf("invalid archive filter pattern %q: %w", pattern, err)
		}
	}

	return &entryFilter{
		include: append([]string(nil), include...),
		exclude: append([]string(nil), exclude...),
	}, nil
}

func (f *entryFilter) matches(name string) bool {
	included := len(f.include) == 0 || matchAny(f.include, name)
	if !included {
		return false
	}

	return !matchAny(f.exclude, name)
}

func matchAny(patterns []string, name string) bool {
	for _, pattern := range patterns {
		matched, _ := path.Match(pattern, name)
		if matched {
			return true
		}
	}

	return false
}
