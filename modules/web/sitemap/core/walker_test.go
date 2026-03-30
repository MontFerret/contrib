package core

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCollectURLs(t *testing.T) {
	t.Run("recursive false skips sitemap indexes", func(t *testing.T) {
		var server *httptest.Server
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/index.xml":
				_, _ = fmt.Fprintf(w, `<sitemapindex><sitemap><loc>%s/posts.xml</loc></sitemap></sitemapindex>`, server.URL)
			case "/posts.xml":
				_, _ = fmt.Fprintf(w, `<urlset><url><loc>%s/a</loc></url></urlset>`, server.URL)
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		opts := DefaultOptions()
		opts.Recursive = false

		entries, err := CollectURLs(t.Context(), server.URL+"/index.xml", opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(entries) != 0 {
			t.Fatalf("expected no entries, got %+v", entries)
		}
	})

	t.Run("dedupes URL entries and preserves first seen source", func(t *testing.T) {
		var server *httptest.Server
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/index.xml":
				_, _ = fmt.Fprintf(w, `<sitemapindex>
					<sitemap><loc>%s/a.xml</loc></sitemap>
					<sitemap><loc>%s/b.xml</loc></sitemap>
				</sitemapindex>`, server.URL, server.URL)
			case "/a.xml":
				_, _ = fmt.Fprintf(w, `<urlset>
					<url><loc>%s/one</loc></url>
					<url><loc>%s/two</loc></url>
				</urlset>`, server.URL, server.URL)
			case "/b.xml":
				_, _ = fmt.Fprintf(w, `<urlset>
					<url><loc>%s/two</loc></url>
					<url><loc>%s/three</loc></url>
				</urlset>`, server.URL, server.URL)
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		entries, err := CollectURLs(t.Context(), server.URL+"/index.xml", DefaultOptions())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(entries) != 3 {
			t.Fatalf("expected 3 entries, got %d", len(entries))
		}

		if entries[0].Loc != server.URL+"/one" || entries[0].Source != server.URL+"/a.xml" {
			t.Fatalf("unexpected first entry %+v", entries[0])
		}

		if entries[1].Loc != server.URL+"/two" || entries[1].Source != server.URL+"/a.xml" {
			t.Fatalf("unexpected second entry %+v", entries[1])
		}

		if entries[2].Loc != server.URL+"/three" || entries[2].Source != server.URL+"/b.xml" {
			t.Fatalf("unexpected third entry %+v", entries[2])
		}

		opts := DefaultOptions()
		opts.Dedupe = false

		allEntries, err := CollectURLs(t.Context(), server.URL+"/index.xml", opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(allEntries) != 4 {
			t.Fatalf("expected 4 entries without dedupe, got %d", len(allEntries))
		}
	})

	t.Run("prevents cycles without dedupe", func(t *testing.T) {
		var server *httptest.Server
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/index.xml":
				_, _ = fmt.Fprintf(w, `<sitemapindex>
					<sitemap><loc>%s/child.xml</loc></sitemap>
					<sitemap><loc>%s/leaf.xml</loc></sitemap>
				</sitemapindex>`, server.URL, server.URL)
			case "/child.xml":
				_, _ = fmt.Fprintf(w, `<sitemapindex>
					<sitemap><loc>%s/index.xml</loc></sitemap>
					<sitemap><loc>%s/leaf.xml</loc></sitemap>
				</sitemapindex>`, server.URL, server.URL)
			case "/leaf.xml":
				_, _ = fmt.Fprintf(w, `<urlset><url><loc>%s/page</loc></url></urlset>`, server.URL)
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		opts := DefaultOptions()
		opts.Dedupe = false

		entries, err := CollectURLs(t.Context(), server.URL+"/index.xml", opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(entries) != 2 {
			t.Fatalf("expected 2 entries from repeated leaf sitemap, got %d", len(entries))
		}
	})

	t.Run("returns expand errors when max depth is exceeded", func(t *testing.T) {
		var server *httptest.Server
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/index.xml":
				_, _ = fmt.Fprintf(w, `<sitemapindex><sitemap><loc>%s/child.xml</loc></sitemap></sitemapindex>`, server.URL)
			case "/child.xml":
				_, _ = fmt.Fprintf(w, `<sitemapindex><sitemap><loc>%s/leaf.xml</loc></sitemap></sitemapindex>`, server.URL)
			case "/leaf.xml":
				_, _ = fmt.Fprintf(w, `<urlset><url><loc>%s/page</loc></url></urlset>`, server.URL)
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		opts := DefaultOptions()
		opts.MaxDepth = 1

		_, err := CollectURLs(t.Context(), server.URL+"/index.xml", opts)
		if err == nil {
			t.Fatal("expected error")
		}

		assertStageError(t, err, StageExpand, server.URL+"/child.xml")
	})
}
