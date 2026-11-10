package network

import (
	"github.com/gobwas/glob"
	"github.com/mafredri/cdp/protocol/network"
)

type (
	ResourceFilter struct {
		URL          *glob.Pattern
		ResourceType string
	}

	// InterceptorFilter blocks matching resources or requests outside its allowed origins.
	// Its compiled configuration is immutable and may be shared between interceptors.
	InterceptorFilter struct {
		resources []ResourceFilter
		origins   map[origin]struct{}
	}
)

// NewInterceptorFilter validates and snapshots a resource/origin filter.
func NewInterceptorFilter(filter *Filter) (*InterceptorFilter, error) {
	interFilter := new(InterceptorFilter)

	if filter.AllowedOrigins != nil {
		origins, err := compileOrigins(filter.AllowedOrigins)
		if err != nil {
			return nil, err
		}

		interFilter.origins = origins
	}

	interFilter.resources = make([]ResourceFilter, 0, len(filter.Patterns))

	for _, pattern := range filter.Patterns {
		rf := ResourceFilter{
			ResourceType: pattern.Type,
		}

		if pattern.URL != "" {
			p, err := glob.Compile(pattern.URL)
			if err != nil {
				return nil, err
			}

			rf.URL = p
		}

		if rf.ResourceType != "" && rf.URL != nil {
			interFilter.resources = append(interFilter.resources, rf)
		}
	}

	return interFilter, nil
}

// Filter reports whether the request must be blocked.
func (f *InterceptorFilter) Filter(rt network.ResourceType, req network.Request) bool {
	if f.origins != nil && req.URL != BlankPageURL {
		parsed, err := parseOrigin(req.URL, false)
		if err != nil {
			return true
		}

		if _, allowed := f.origins[parsed]; !allowed {
			return true
		}
	}

	var result bool

	for _, pattern := range f.resources {
		if pattern.ResourceType != "" && pattern.URL != nil {
			result = string(rt) == pattern.ResourceType && pattern.URL.Match(req.URL)
		} else if pattern.ResourceType != "" {
			result = string(rt) == pattern.ResourceType
		} else if pattern.URL != nil {
			result = pattern.URL.Match(req.URL)
		}

		if result {
			break
		}
	}

	return result
}
