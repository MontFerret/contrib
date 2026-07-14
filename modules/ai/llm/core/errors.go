package core

// Error is a sanitized AI::LLM error with a stable code.
type Error struct {
	Code      ErrorCode
	Message   string
	operation string
}

// NewError creates a sanitized typed error.
func NewError(code ErrorCode, message string) *Error {
	return &Error{Code: code, Message: message}
}

func (e *Error) Error() string {
	if e.operation == "" {
		return string(e.Code) + ": " + e.Message
	}

	return string(e.Code) + ": " + e.operation + ": " + e.Message
}

// WithOperation returns a copy annotated with the public operation name.
func (e *Error) WithOperation(operation string) *Error {
	if e == nil {
		return nil
	}

	cp := *e
	cp.operation = operation

	return &cp
}
