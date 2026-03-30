package core

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetch(t *testing.T) {
	t.Run("rejects invalid URL", func(t *testing.T) {
		_, err := Fetch(t.Context(), "://bad", DefaultOptions())
		if err == nil {
			t.Fatal("expected error")
		}

		assertStageError(t, err, StageFetch, "://bad")
	})

	t.Run("rejects unsupported scheme", func(t *testing.T) {
		_, err := Fetch(t.Context(), "ftp://example.com/sitemap.xml", DefaultOptions())
		if err == nil {
			t.Fatal("expected error")
		}

		assertStageError(t, err, StageFetch, "ftp://example.com/sitemap.xml")
	})

	t.Run("rejects non success responses", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "missing", http.StatusNotFound)
		}))
		defer server.Close()

		_, err := Fetch(t.Context(), server.URL+"/missing.xml", DefaultOptions())
		if err == nil {
			t.Fatal("expected error")
		}

		assertStageError(t, err, StageFetch, server.URL+"/missing.xml")
	})

	t.Run("sends headers and parses response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("User-Agent"); got != "Ferret/2" {
				t.Fatalf("expected User-Agent header, got %q", got)
			}

			_, _ = fmt.Fprintf(w, `<urlset><url><loc>%s/page</loc></url></urlset>`, serverURL(r))
		}))
		defer server.Close()

		opts := DefaultOptions()
		opts.Headers = map[string]string{
			"User-Agent": "Ferret/2",
		}

		doc, err := Fetch(t.Context(), server.URL+"/sitemap.xml", opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(doc.URLs) != 1 || doc.URLs[0].Loc != server.URL+"/page" {
			t.Fatalf("unexpected document %+v", doc)
		}
	})

	t.Run("honors timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(50 * time.Millisecond)
			_, _ = w.Write([]byte(`<urlset/>`))
		}))
		defer server.Close()

		opts := DefaultOptions()
		opts.Timeout = 5 * time.Millisecond

		_, err := Fetch(t.Context(), server.URL+"/slow.xml", opts)
		if err == nil {
			t.Fatal("expected error")
		}

		assertStageError(t, err, StageFetch, server.URL+"/slow.xml")
	})
}

func serverURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	return scheme + "://" + r.Host
}
