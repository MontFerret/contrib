package network

import (
	"fmt"
	"testing"

	"github.com/mafredri/cdp/rpcc"
)

func TestIsUnavailableResponseBodyError(t *testing.T) {
	tests := []struct {
		err  error
		name string
		want bool
	}{
		{name: "unavailable", err: &rpcc.ResponseError{Code: -32000, Message: "No resource with given identifier found"}, want: true},
		{name: "evicted wrapped", err: fmt.Errorf("get body: %w", &rpcc.ResponseError{Code: -32000, Message: "Request content was evicted from inspector cache"}), want: true},
		{name: "unknown minus 32000", err: &rpcc.ResponseError{Code: -32000, Message: "Internal browser failure"}},
		{name: "wrong code", err: &rpcc.ResponseError{Code: -32602, Message: "No resource with given identifier found"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUnavailableResponseBodyError(tt.err); got != tt.want {
				t.Fatalf("isUnavailableResponseBodyError() = %v, want %v", got, tt.want)
			}
		})
	}
}
