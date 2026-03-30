package core

import (
	"errors"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	t.Run("parses groups metadata and preserves order", func(t *testing.T) {
		doc, err := Parse(strings.TrimSpace(`
			# top-level metadata
			Disallow: /ignored-before-group
			Sitemap: https://example.com/sitemap.xml
			Host: example.com

			User-agent: FooBot
			User-agent: BarBot
			Allow: /public
			Disallow: /private
			Crawl-delay: 5

			User-agent: *
			Disallow:
			Unknown: ignored
			Disallow: /tmp

			User-agent: FooBot
			Disallow: /extra
		`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(doc.Groups) != 3 {
			t.Fatalf("expected 3 groups, got %d", len(doc.Groups))
		}

		if got := doc.Groups[0].UserAgents; !slicesEqual(got, []string{"FooBot", "BarBot"}) {
			t.Fatalf("unexpected first group userAgents %v", got)
		}

		if got := doc.Groups[0].Allow; !slicesEqual(got, []string{"/public"}) {
			t.Fatalf("unexpected first group allow %v", got)
		}

		if got := doc.Groups[0].Disallow; !slicesEqual(got, []string{"/private"}) {
			t.Fatalf("unexpected first group disallow %v", got)
		}

		if doc.Groups[0].CrawlDelay == nil || *doc.Groups[0].CrawlDelay != 5 {
			t.Fatalf("unexpected first group crawlDelay %v", doc.Groups[0].CrawlDelay)
		}

		if got := doc.Groups[1].UserAgents; !slicesEqual(got, []string{"*"}) {
			t.Fatalf("unexpected second group userAgents %v", got)
		}

		if got := doc.Groups[1].Disallow; !slicesEqual(got, []string{"/tmp"}) {
			t.Fatalf("unexpected second group disallow %v", got)
		}

		if got := doc.Groups[2].Disallow; !slicesEqual(got, []string{"/extra"}) {
			t.Fatalf("unexpected third group disallow %v", got)
		}

		if got := doc.Sitemaps; !slicesEqual(got, []string{"https://example.com/sitemap.xml"}) {
			t.Fatalf("unexpected sitemaps %v", got)
		}

		if doc.Host == nil || *doc.Host != "example.com" {
			t.Fatalf("unexpected host %v", doc.Host)
		}
	})

	t.Run("ignores blank lines comments and empty metadata", func(t *testing.T) {
		doc, err := Parse(strings.TrimSpace(`
			# comment
			Sitemap:

			User-agent: *
			Allow: /public # inline comment
			Host:
		`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(doc.Sitemaps) != 0 {
			t.Fatalf("expected no sitemaps, got %v", doc.Sitemaps)
		}

		if doc.Host != nil {
			t.Fatalf("expected nil host, got %v", *doc.Host)
		}

		if got := doc.Groups[0].Allow; !slicesEqual(got, []string{"/public"}) {
			t.Fatalf("unexpected allow %v", got)
		}
	})

	t.Run("rejects empty user-agent", func(t *testing.T) {
		_, err := Parse("User-agent:   ")
		if err == nil {
			t.Fatal("expected error")
		}

		var parseErr *Error
		if !errors.As(err, &parseErr) || parseErr.Stage != StageParse {
			t.Fatalf("expected parse error, got %v", err)
		}
	})

	t.Run("rejects non-numeric crawl-delay", func(t *testing.T) {
		_, err := Parse(strings.TrimSpace(`
			User-agent: *
			Crawl-delay: slow
		`))
		if err == nil {
			t.Fatal("expected error")
		}

		var parseErr *Error
		if !errors.As(err, &parseErr) || parseErr.Stage != StageParse {
			t.Fatalf("expected parse error, got %v", err)
		}
	})
}

func slicesEqual[T comparable](got, want []T) bool {
	if len(got) != len(want) {
		return false
	}

	for idx := range got {
		if got[idx] != want[idx] {
			return false
		}
	}

	return true
}
