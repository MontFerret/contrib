package core

import (
	"context"
	"strings"
	"testing"

	xmlcore "github.com/MontFerret/contrib/modules/xml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestParse(t *testing.T) {
	t.Run("parses urlset with namespaces and unknown tags", func(t *testing.T) {
		doc, err := Parse(strings.NewReader(`
			<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
			  <url>
			    <loc>https://example.com/a</loc>
			    <lastmod>2026-03-01T12:00:00Z</lastmod>
			    <changefreq>weekly</changefreq>
			    <priority>0.8</priority>
			    <image:image>
			      <image:loc>https://example.com/a.png</image:loc>
			    </image:image>
			  </url>
			  <url>
			    <loc>https://example.com/b</loc>
			  </url>
			</urlset>
		`), "https://example.com/sitemap.xml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if doc.Type != TypeURLSet {
			t.Fatalf("expected type %q, got %q", TypeURLSet, doc.Type)
		}

		if len(doc.URLs) != 2 {
			t.Fatalf("expected 2 URLs, got %d", len(doc.URLs))
		}

		first := doc.URLs[0]
		if first.Loc != "https://example.com/a" {
			t.Fatalf("unexpected first loc %q", first.Loc)
		}

		if first.LastMod != "2026-03-01T12:00:00Z" {
			t.Fatalf("unexpected lastmod %q", first.LastMod)
		}

		if first.ChangeFreq != "weekly" {
			t.Fatalf("unexpected changefreq %q", first.ChangeFreq)
		}

		if first.Priority == nil || *first.Priority != 0.8 {
			t.Fatalf("expected priority 0.8, got %v", first.Priority)
		}

		if first.Source != "https://example.com/sitemap.xml" {
			t.Fatalf("unexpected source %q", first.Source)
		}

		second := doc.URLs[1]
		if second.Loc != "https://example.com/b" {
			t.Fatalf("unexpected second loc %q", second.Loc)
		}

		if second.LastMod != "" || second.ChangeFreq != "" || second.Priority != nil {
			t.Fatalf("expected optional fields to be empty, got %+v", second)
		}
	})

	t.Run("parses sitemapindex with prefixed namespaces", func(t *testing.T) {
		doc, err := Parse(strings.NewReader(`
			<sm:sitemapindex xmlns:sm="http://www.sitemaps.org/schemas/sitemap/0.9">
			  <sm:sitemap>
			    <sm:loc>https://example.com/posts.xml</sm:loc>
			    <sm:lastmod>2026-03-02</sm:lastmod>
			  </sm:sitemap>
			</sm:sitemapindex>
		`), "https://example.com/index.xml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if doc.Type != TypeSitemapIndex {
			t.Fatalf("expected type %q, got %q", TypeSitemapIndex, doc.Type)
		}

		if len(doc.Sitemaps) != 1 {
			t.Fatalf("expected 1 sitemap, got %d", len(doc.Sitemaps))
		}

		ref := doc.Sitemaps[0]
		if ref.Loc != "https://example.com/posts.xml" {
			t.Fatalf("unexpected loc %q", ref.Loc)
		}

		if ref.LastMod != "2026-03-02" {
			t.Fatalf("unexpected lastmod %q", ref.LastMod)
		}
	})

	t.Run("rejects unsupported roots", func(t *testing.T) {
		_, err := Parse(strings.NewReader(`<feed></feed>`), "https://example.com/feed.xml")
		if err == nil {
			t.Fatal("expected error")
		}

		assertStageError(t, err, StageParse, "https://example.com/feed.xml")
	})

	t.Run("rejects malformed xml", func(t *testing.T) {
		_, err := Parse(strings.NewReader(`<urlset><url><loc>https://example.com</url></urlset>`), "https://example.com/bad.xml")
		if err == nil {
			t.Fatal("expected error")
		}

		assertStageError(t, err, StageParse, "https://example.com/bad.xml")
	})

	t.Run("xml core iterator remains compatible with sitemap interpreter", func(t *testing.T) {
		iter, err := xmlcore.NewDecodeIterator(runtime.NewString(`
			<sm:sitemapindex xmlns:sm="http://www.sitemaps.org/schemas/sitemap/0.9">
			  <sm:sitemap>
			    <sm:loc>https://example.com/posts.xml</sm:loc>
			  </sm:sitemap>
			</sm:sitemapindex>
		`))
		if err != nil {
			t.Fatalf("unexpected iterator error: %v", err)
		}

		doc, err := parseIterator(context.Background(), iter, "https://example.com/index.xml")
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}

		if doc.Type != TypeSitemapIndex {
			t.Fatalf("expected type %q, got %q", TypeSitemapIndex, doc.Type)
		}

		if len(doc.Sitemaps) != 1 || doc.Sitemaps[0].Loc != "https://example.com/posts.xml" {
			t.Fatalf("unexpected sitemaps %+v", doc.Sitemaps)
		}
	})
}
