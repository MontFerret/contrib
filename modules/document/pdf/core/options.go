package core

const defaultMaxBufferSize int64 = 64 * 1024 * 1024

// OpenOptions controls how PDF sources are adapted for the parser.
type OpenOptions struct {
	MaxBufferSize int64
}

// DefaultOpenOptions returns the default PDF open policy.
func DefaultOpenOptions() OpenOptions {
	return OpenOptions{
		MaxBufferSize: defaultMaxBufferSize,
	}
}

func (o OpenOptions) normalize() (OpenOptions, error) {
	if o.MaxBufferSize == 0 {
		o.MaxBufferSize = defaultMaxBufferSize
	}
	if o.MaxBufferSize < 0 {
		return o, OperationErrorf("OPEN", "maximum buffer size must be non-negative, got %d", o.MaxBufferSize)
	}

	return o, nil
}
