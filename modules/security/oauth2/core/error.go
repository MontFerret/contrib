package core

import (
	"fmt"
	"io"
)

type (
	// Error is an OAuth protocol error returned by a token endpoint.
	Error struct {
		Operation   string
		Code        string
		Description string
		URI         string
		StatusCode  int
	}

	safeOAuthError struct {
		Type        string `json:"type"`
		Operation   string `json:"operation,omitempty"`
		Code        string `json:"code"`
		Description string `json:"description,omitempty"`
		URI         string `json:"uri,omitempty"`
		StatusCode  int    `json:"statusCode,omitempty"`
	}
)

// Error returns the OAuth error without request or response bodies.
func (e *Error) Error() string {
	if e == nil {
		return "oauth2 token request failed"
	}

	operation := e.Operation
	if operation == "" {
		operation = "token request"
	}

	message := "oauth2 " + operation
	if e.Code != "" {
		message += ": " + e.Code
	}

	if e.Description != "" {
		message += ": " + e.Description
	}

	return message
}

// String returns the same secret-free representation as Error.
func (e *Error) String() string {
	return e.Error()
}

// Format implements fmt.Formatter without exposing raw response data.
func (e *Error) Format(state fmt.State, verb rune) {
	value := e.Error()
	if verb == 'q' {
		value = fmt.Sprintf("%q", value)
	}

	_, _ = io.WriteString(state, value)
}

// MarshalJSON serializes only normalized OAuth error fields.
func (e *Error) MarshalJSON() ([]byte, error) {
	if e == nil {
		return []byte("null"), nil
	}

	return marshalSafeJSON(safeOAuthError{
		Type:        "oauth2.Error",
		Operation:   e.Operation,
		Code:        e.Code,
		Description: e.Description,
		URI:         e.URI,
		StatusCode:  e.StatusCode,
	})
}
