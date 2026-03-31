package core

import (
	"context"
	"net/url"
	"strings"
	"testing"
)

func TestExtractorFromContextReturnsInjectedExtractor(t *testing.T) {
	extractor := NewExtractor()
	ctx := WithExtractor(context.Background(), extractor)

	got := ExtractorFromContext(ctx)
	if got == nil {
		t.Fatal("expected extractor from context")
	}

	if got != extractor {
		t.Fatal("expected injected extractor instance")
	}
}

func TestExtractorFallbackUsesNewConverterWhenMissing(t *testing.T) {
	first := (*Extractor)(nil).markdownConverterOrNew()
	second := (*Extractor)(nil).markdownConverterOrNew()

	if first == nil || second == nil {
		t.Fatal("expected markdown converters")
	}

	if first == second {
		t.Fatal("expected separate fallback converter instances")
	}
}

func TestRenderMarkdownUsesPerCallDomainWithoutLeakage(t *testing.T) {
	firstBaseURL, err := url.Parse("https://first.example/posts/alpha")
	if err != nil {
		t.Fatalf("unexpected first base url parse error: %v", err)
	}

	secondBaseURL, err := url.Parse("https://second.example/docs/beta")
	if err != nil {
		t.Fatalf("unexpected second base url parse error: %v", err)
	}

	fragment := `<p><a href="/story">Story</a> <img src="/hero.jpg" alt="Lead" /></p>`
	extractor := NewExtractor()

	first := extractor.renderMarkdown(fragment, firstBaseURL)
	if first == nil {
		t.Fatal("expected markdown for first conversion")
	}

	if !strings.Contains(*first, "https://first.example/story") || !strings.Contains(*first, "https://first.example/hero.jpg") {
		t.Fatalf("unexpected first markdown %q", *first)
	}

	second := extractor.renderMarkdown(fragment, secondBaseURL)
	if second == nil {
		t.Fatal("expected markdown for second conversion")
	}

	if !strings.Contains(*second, "https://second.example/story") || !strings.Contains(*second, "https://second.example/hero.jpg") {
		t.Fatalf("unexpected second markdown %q", *second)
	}

	if strings.Contains(*second, "https://first.example/") {
		t.Fatalf("expected second markdown to avoid leaked first domain, got %q", *second)
	}

	withoutBase := extractor.renderMarkdown(fragment, nil)
	if withoutBase == nil {
		t.Fatal("expected markdown without base url")
	}

	if !strings.Contains(*withoutBase, "(/story)") || !strings.Contains(*withoutBase, "(/hero.jpg)") {
		t.Fatalf("unexpected markdown without base %q", *withoutBase)
	}

	if strings.Contains(*withoutBase, "https://first.example/") || strings.Contains(*withoutBase, "https://second.example/") {
		t.Fatalf("expected markdown without base to avoid leaked domains, got %q", *withoutBase)
	}
}
