package drivers

import (
	"errors"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestNormalizeInitScript(t *testing.T) {
	tests := []struct {
		name       string
		script     *InitScript
		wantTiming InitScriptTiming
		wantErr    bool
	}{
		{name: "absent"},
		{name: "default timing", script: &InitScript{Source: "window.ready = true"}, wantTiming: InitScriptAfterNavigation},
		{name: "before document", script: &InitScript{Source: "window.ready = true", Timing: InitScriptBeforeDocument}, wantTiming: InitScriptBeforeDocument},
		{name: "blank source", script: &InitScript{Source: " \t"}, wantErr: true},
		{name: "invalid timing", script: &InitScript{Source: "true", Timing: "duringNavigation"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeInitScript(tt.script)
			if tt.wantErr {
				if !errors.Is(err, runtime.ErrInvalidArgument) {
					t.Fatalf("expected invalid argument, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.script == nil {
				if got != nil {
					t.Fatalf("expected nil, got %#v", got)
				}
				return
			}
			if got.Timing != tt.wantTiming {
				t.Fatalf("timing = %q, want %q", got.Timing, tt.wantTiming)
			}
			if got == tt.script {
				t.Fatal("normalization must not mutate the caller's configuration")
			}
		})
	}
}
