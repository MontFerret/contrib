package core

func normalizeSessionOptions(options SessionOptions) (SessionOptions, error) {
	if options.Context.Mode == "" {
		options.Context.Mode = "local"
	}
	if options.Context.Mode != "local" {
		return SessionOptions{}, NewError(ErrInvalidOptions, "session context mode must be local")
	}

	if options.Context.Overflow == "" {
		options.Context.Overflow = "error"
	}
	if options.Context.Overflow != "error" {
		return SessionOptions{}, NewError(ErrInvalidOptions, "session context overflow must be error")
	}

	if options.Context.MaxTokens < 0 {
		return SessionOptions{}, NewError(ErrInvalidOptions, "session maxTokens must be positive")
	}
	if options.Context.ReserveOutputTokens < 0 {
		return SessionOptions{}, NewError(ErrInvalidOptions, "session reserveOutputTokens must be nonnegative")
	}
	if options.Context.MaxTokens > 0 && options.Context.ReserveOutputTokens >= options.Context.MaxTokens {
		return SessionOptions{}, NewError(ErrInvalidOptions, "session output token reserve must be smaller than maxTokens")
	}

	return options, nil
}
