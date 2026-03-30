package lib

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestStream(t *testing.T) {
	t.Run("returns proxied iterator with sequential keys", func(t *testing.T) {
		var server *httptest.Server
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/sitemap.xml":
				_, _ = fmt.Fprintf(w, `<urlset><url><loc>%s/one</loc></url></urlset>`, server.URL)
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		value, err := Stream(t.Context(), runtime.NewString(server.URL+"/sitemap.xml"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		iter := mustIterate(t, t.Context(), value)
		entry, key, err := iter.Next(t.Context())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if key.(runtime.Int) != 1 {
			t.Fatalf("expected key 1, got %v", key)
		}

		obj := mustRuntimeObject(t, entry)
		if got := mustObjectField(t, t.Context(), obj, "loc").String(); got != server.URL+"/one" {
			t.Fatalf("unexpected loc %q", got)
		}
	})

	t.Run("defers expansion errors until iteration", func(t *testing.T) {
		var server *httptest.Server
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/index.xml":
				_, _ = fmt.Fprintf(w, `<sitemapindex>
					<sitemap><loc>%s/good.xml</loc></sitemap>
					<sitemap><loc>%s/bad.xml</loc></sitemap>
				</sitemapindex>`, server.URL, server.URL)
			case "/good.xml":
				_, _ = fmt.Fprintf(w, `<urlset><url><loc>%s/good</loc></url></urlset>`, server.URL)
			case "/bad.xml":
				http.Error(w, "bad", http.StatusInternalServerError)
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		value, err := Stream(t.Context(), runtime.NewString(server.URL+"/index.xml"))
		if err != nil {
			t.Fatalf("unexpected stream error: %v", err)
		}

		iter := mustIterate(t, t.Context(), value)
		if _, _, err := iter.Next(t.Context()); err != nil {
			t.Fatalf("unexpected first error: %v", err)
		}

		if _, _, err := iter.Next(t.Context()); err == nil {
			t.Fatal("expected deferred iteration error")
		}
	})
}
