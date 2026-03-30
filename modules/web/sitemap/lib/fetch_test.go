package lib

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != "Ferret/2" {
			t.Fatalf("expected User-Agent header, got %q", got)
		}

		_, _ = fmt.Fprintf(w, `<urlset><url><loc>%s/page</loc></url></urlset>`, serverURL(r))
	}))
	defer server.Close()

	result, err := Fetch(t.Context(),
		runtime.NewString(server.URL+"/sitemap.xml"),
		runtime.NewObjectWith(map[string]runtime.Value{
			"headers": runtime.NewObjectWith(map[string]runtime.Value{
				"User-Agent": runtime.NewString("Ferret/2"),
			}),
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj := mustRuntimeObject(t, result)
	if got := mustObjectField(t, t.Context(), obj, "type").String(); got != "urlset" {
		t.Fatalf("unexpected type %q", got)
	}

	urls := mustRuntimeArray(t, mustObjectField(t, t.Context(), obj, "urls"))
	if length, err := urls.Length(t.Context()); err != nil || length != 1 {
		t.Fatalf("unexpected urls length %v %v", length, err)
	}
}

func TestFetchRejectsInvalidOptions(t *testing.T) {
	_, err := Fetch(t.Context(), runtime.NewString("https://example.com/sitemap.xml"), runtime.NewInt(1))
	if err == nil {
		t.Fatal("expected error")
	}
}

func serverURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	return scheme + "://" + r.Host
}
