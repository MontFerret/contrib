package lib

import (
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestParse(t *testing.T) {
	result, err := Parse(t.Context(), runtime.NewString(`
		Sitemap: https://example.com/sitemap.xml
		User-agent: *
		Allow: /public
		Disallow: /admin
		Crawl-delay: 2
	`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj := mustRuntimeObject(t, result)
	groups := mustRuntimeArray(t, mustObjectField(t, t.Context(), obj, "groups"))
	if length, err := groups.Length(t.Context()); err != nil || length != 1 {
		t.Fatalf("unexpected groups length %v %v", length, err)
	}

	groupValue, err := groups.At(t.Context(), 0)
	if err != nil {
		t.Fatalf("unexpected group get error: %v", err)
	}

	group := mustRuntimeObject(t, groupValue)
	if got := mustArrayStrings(t, t.Context(), mustRuntimeArray(t, mustObjectField(t, t.Context(), group, "userAgents"))); len(got) != 1 || got[0] != "*" {
		t.Fatalf("unexpected userAgents %v", got)
	}

	if got := mustArrayStrings(t, t.Context(), mustRuntimeArray(t, mustObjectField(t, t.Context(), group, "allow"))); len(got) != 1 || got[0] != "/public" {
		t.Fatalf("unexpected allow %v", got)
	}

	if got := mustArrayStrings(t, t.Context(), mustRuntimeArray(t, mustObjectField(t, t.Context(), group, "disallow"))); len(got) != 1 || got[0] != "/admin" {
		t.Fatalf("unexpected disallow %v", got)
	}

	crawlDelay := mustObjectField(t, t.Context(), group, "crawlDelay")
	if crawlDelay.String() != "2" {
		t.Fatalf("unexpected crawlDelay %v", crawlDelay)
	}

	sitemaps := mustRuntimeArray(t, mustObjectField(t, t.Context(), obj, "sitemaps"))
	if got := mustArrayStrings(t, t.Context(), sitemaps); len(got) != 1 || got[0] != "https://example.com/sitemap.xml" {
		t.Fatalf("unexpected sitemaps %v", got)
	}
}

func TestParseRejectsInvalidArgs(t *testing.T) {
	if _, err := Parse(t.Context()); err == nil {
		t.Fatal("expected error")
	}

	if _, err := Parse(t.Context(), runtime.NewInt(1)); err == nil {
		t.Fatal("expected error")
	}
}
