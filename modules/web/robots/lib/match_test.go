package lib

import (
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestAllowsAndMatch(t *testing.T) {
	parsed, err := Parse(t.Context(), runtime.NewString(`
		User-agent: *
		Disallow: /admin
		Allow: /admin/public
		Sitemap: https://example.com/sitemap.xml
	`))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	allowed, err := Allows(t.Context(), parsed, runtime.NewString("/admin/public/page"), runtime.NewString("FerretBot"))
	if err != nil {
		t.Fatalf("unexpected allows error: %v", err)
	}

	if allowed != runtime.True {
		t.Fatalf("expected allowed result, got %v", allowed)
	}

	matched, err := Match(t.Context(), parsed, runtime.NewString("/admin/page"))
	if err != nil {
		t.Fatalf("unexpected match error: %v", err)
	}

	obj := mustRuntimeObject(t, matched)
	if got := mustObjectField(t, t.Context(), obj, "allowed"); got != runtime.False {
		t.Fatalf("unexpected allowed field %v", got)
	}

	if got := mustObjectField(t, t.Context(), obj, "directive").String(); got != "disallow" {
		t.Fatalf("unexpected directive %q", got)
	}

	if got := mustObjectField(t, t.Context(), obj, "pattern").String(); got != "/admin" {
		t.Fatalf("unexpected pattern %q", got)
	}

	if got := mustObjectField(t, t.Context(), obj, "userAgent").String(); got != "*" {
		t.Fatalf("unexpected userAgent %q", got)
	}

	fallbackMatched, err := Match(t.Context(), parsed, runtime.NewString("/admin/page"), runtime.NewString("FerretBot"))
	if err != nil {
		t.Fatalf("unexpected fallback match error: %v", err)
	}

	fallbackObj := mustRuntimeObject(t, fallbackMatched)
	if got := mustObjectField(t, t.Context(), fallbackObj, "userAgent").String(); got != "*" {
		t.Fatalf("expected wildcard fallback userAgent %q, got %q", "*", got)
	}

	exactParsed, err := Parse(t.Context(), runtime.NewString(`
		User-agent: *
		Disallow: /admin
		User-agent: FerretBot
		Allow: /admin
	`))
	if err != nil {
		t.Fatalf("unexpected exact parse error: %v", err)
	}

	exactMatched, err := Match(t.Context(), exactParsed, runtime.NewString("/admin/page"), runtime.NewString("FerretBot"))
	if err != nil {
		t.Fatalf("unexpected exact match error: %v", err)
	}

	exactObj := mustRuntimeObject(t, exactMatched)
	if got := mustObjectField(t, t.Context(), exactObj, "userAgent").String(); got != "FerretBot" {
		t.Fatalf("expected exact-match userAgent %q, got %q", "FerretBot", got)
	}
}

func TestSitemapsAndManualObject(t *testing.T) {
	robots := runtime.NewObjectWith(map[string]runtime.Value{
		"groups": runtime.NewArrayOf([]runtime.Value{
			runtime.NewObjectWith(map[string]runtime.Value{
				"userAgents": runtime.NewArrayOf([]runtime.Value{runtime.NewString("ManualBot")}),
				"allow":      runtime.NewArrayOf([]runtime.Value{runtime.NewString("/public")}),
				"disallow":   runtime.NewArrayOf([]runtime.Value{runtime.NewString("/private")}),
				"crawlDelay": runtime.NewInt(3),
			}),
		}),
		"sitemaps": runtime.NewArrayOf([]runtime.Value{
			runtime.NewString("https://example.com/sitemap.xml"),
			runtime.NewString("https://example.com/news.xml"),
		}),
		"host": runtime.None,
	})

	allowed, err := Allows(t.Context(), robots, runtime.NewString("/public/page"), runtime.NewString("manualbot"))
	if err != nil {
		t.Fatalf("unexpected allows error: %v", err)
	}

	if allowed != runtime.True {
		t.Fatalf("expected allowed result, got %v", allowed)
	}

	sitemaps, err := Sitemaps(t.Context(), robots)
	if err != nil {
		t.Fatalf("unexpected sitemaps error: %v", err)
	}

	values := mustArrayStrings(t, t.Context(), mustRuntimeArray(t, sitemaps))
	if len(values) != 2 || values[0] != "https://example.com/sitemap.xml" || values[1] != "https://example.com/news.xml" {
		t.Fatalf("unexpected sitemaps %v", values)
	}
}

func TestDecodeRejectsMalformedObjects(t *testing.T) {
	invalid := runtime.NewObjectWith(map[string]runtime.Value{
		"groups": runtime.NewString("bad"),
	})

	if _, err := Allows(t.Context(), invalid, runtime.NewString("/"), runtime.NewString("Bot")); err == nil {
		t.Fatal("expected error")
	}

	if _, err := Match(t.Context(), invalid, runtime.NewString("/")); err == nil {
		t.Fatal("expected error")
	}

	if _, err := Sitemaps(t.Context(), invalid); err == nil {
		t.Fatal("expected error")
	}
}
