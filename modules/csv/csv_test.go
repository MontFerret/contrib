package csv

import (
	"strings"
	"testing"

	ferret "github.com/MontFerret/ferret/v2"
)

func TestDuplicateModuleRegistrationReportsModuleContext(t *testing.T) {
	t.Parallel()

	engine, err := ferret.New(ferret.WithModules(New(), New()))
	if engine != nil {
		t.Cleanup(func() {
			if closeErr := engine.Close(); closeErr != nil {
				t.Errorf("unexpected engine close error: %v", closeErr)
			}
		})
	}
	if err == nil {
		t.Fatal("expected duplicate module registration error")
	}
	if !strings.Contains(err.Error(), `module "csv": register`) {
		t.Fatalf("expected module registration context, got %v", err)
	}
}
