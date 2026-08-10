package eval

import (
	"fmt"
	"testing"

	"github.com/mafredri/cdp/rpcc"
)

func TestIsStaleError(t *testing.T) {
	tests := []struct {
		err  error
		name string
		want bool
	}{
		{name: "typed", err: &rpcc.ResponseError{Code: -32000, Message: "Execution context was destroyed."}, want: true},
		{name: "wrapped", err: fmt.Errorf("outer: %w", &rpcc.ResponseError{Code: -32000, Message: "Could not find object with given id"}), want: true},
		{name: "generic minus 32000", err: &rpcc.ResponseError{Code: -32000, Message: "Some other protocol error"}},
		{name: "matching message with other code", err: &rpcc.ResponseError{Code: -32602, Message: "Execution context was destroyed."}},
		{name: "plain error", err: fmt.Errorf("execution context was destroyed")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsStaleError(tt.err); got != tt.want {
				t.Fatalf("IsStaleError() = %v, want %v", got, tt.want)
			}
		})
	}
}
