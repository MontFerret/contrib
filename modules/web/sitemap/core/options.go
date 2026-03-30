package core

import "time"

const (
	DefaultMaxDepth = 8
	DefaultTimeout  = 30 * time.Second
)

// Options configures sitemap fetching and traversal behavior.
type Options struct {
	Headers   map[string]string
	MaxDepth  int
	Timeout   time.Duration
	Recursive bool
	Dedupe    bool
}

// DefaultOptions returns the default sitemap options.
func DefaultOptions() Options {
	return Options{
		Recursive: true,
		Dedupe:    true,
		MaxDepth:  DefaultMaxDepth,
		Timeout:   DefaultTimeout,
		Headers:   map[string]string{},
	}
}
