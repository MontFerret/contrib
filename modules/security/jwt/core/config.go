package core

// Config holds module-level settings passed from registration.
type Config struct {
	MaxTokenSize int
}

func (c Config) maxTokenSize() int {
	if c.MaxTokenSize > 0 {
		return c.MaxTokenSize
	}

	return 64 * 1024
}
