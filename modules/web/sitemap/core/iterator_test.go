package core

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestURLIterator(t *testing.T) {
	t.Run("yields sequential keys and supports close", func(t *testing.T) {
		var server *httptest.Server
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/sitemap.xml":
				_, _ = fmt.Fprintf(w, `<urlset>
					<url><loc>%s/one</loc></url>
					<url><loc>%s/two</loc></url>
				</urlset>`, server.URL, server.URL)
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		iter := NewURLIterator(server.URL+"/sitemap.xml", DefaultOptions())

		first, key, err := iter.Next(t.Context())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if key.(runtime.Int) != 1 {
			t.Fatalf("expected key 1, got %v", key)
		}

		firstObj := mustRuntimeObject(t, first)
		if got := mustObjectField(t, t.Context(), firstObj, "loc").String(); got != server.URL+"/one" {
			t.Fatalf("unexpected loc %q", got)
		}

		_, key, err = iter.Next(t.Context())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if key.(runtime.Int) != 2 {
			t.Fatalf("expected key 2, got %v", key)
		}

		if err := iter.Close(); err != nil {
			t.Fatalf("unexpected close error: %v", err)
		}

		_, _, err = iter.Next(t.Context())
		if err != io.EOF {
			t.Fatalf("expected EOF after close, got %v", err)
		}
	})

	t.Run("defers fetch errors until iteration", func(t *testing.T) {
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

		iter := NewURLIterator(server.URL+"/index.xml", DefaultOptions())

		_, _, err := iter.Next(t.Context())
		if err != nil {
			t.Fatalf("unexpected first error: %v", err)
		}

		_, _, err = iter.Next(t.Context())
		if err == nil {
			t.Fatal("expected deferred error")
		}

		assertStageError(t, err, StageFetch, server.URL+"/bad.xml")
	})
}
