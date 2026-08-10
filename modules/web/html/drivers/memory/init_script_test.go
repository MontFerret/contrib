package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestOpenRejectsInitScript(t *testing.T) {
	driver := New()
	_, err := driver.Open(context.Background(), drivers.Params{
		URL:        "https://example.com",
		InitScript: &drivers.InitScript{Source: "true"},
	})
	if !errors.Is(err, runtime.ErrNotSupported) {
		t.Fatalf("expected not supported error, got %v", err)
	}
}
