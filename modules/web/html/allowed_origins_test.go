package html_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MontFerret/contrib/modules/web/html"
	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/memory"
	"github.com/MontFerret/ferret/v2"
	"github.com/MontFerret/ferret/v2/pkg/sdk/sdktest"
)

func TestDefaultMemoryDriverDocumentAllowlist(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		fmt.Fprint(w, "<title>FQL document</title>")
	}))
	defer server.Close()

	for _, test := range []struct {
		name    string
		origins []string
		denied  bool
	}{
		{name: "listed", origins: []string{server.URL}},
		{name: "empty", denied: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			driver := memory.New(memory.WithOptions(drivers.WithAllowedOrigins(test.origins...)))
			defer driver.Close()

			harness := sdktest.New(t, ferret.WithModules(html.New(html.WithDefaultDriver(driver))))
			before := hits.Load()
			var content *ferret.Content

			output, err := harness.Run(t.Context(), fmt.Sprintf("RETURN DOCUMENT(%q).title", server.URL))
			if err == nil {
				defer output.Close()
				content, err = output.Collect(t.Context())
			}

			if test.denied {
				assert.ErrorContains(t, err, "denied by origin allowlist")
				assert.Equal(t, before, hits.Load())

				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if assert.NotNil(t, content) {
				assert.JSONEq(t, `"FQL document"`, string(content.Data))
			}

			assert.Equal(t, before+1, hits.Load())
		})
	}
}
