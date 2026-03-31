package core

import (
	"context"
	"net/url"
	"strings"
	"testing"
)

func TestMarkdownConverterFromContextReturnsInjectedConverter(t *testing.T) {
	conv := NewMarkdownConverter()
	ctx := WithMarkdownConverter(context.Background(), conv)

	got := MarkdownConverterFromContext(ctx)
	if got == nil {
		t.Fatal("expected markdown converter from context")
	}

	if got != conv {
		t.Fatal("expected injected markdown converter instance")
	}
}

func TestResolveMarkdownConverterReturnsFreshInstanceWithoutContext(t *testing.T) {
	first := resolveMarkdownConverter(context.Background())
	second := resolveMarkdownConverter(context.Background())

	if first == nil || second == nil {
		t.Fatal("expected markdown converter instances")
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
	ctx := WithMarkdownConverter(context.Background(), NewMarkdownConverter())

	first := renderMarkdown(ctx, fragment, firstBaseURL)
	if first == nil {
		t.Fatal("expected markdown for first conversion")
	}

	if !strings.Contains(*first, "https://first.example/story") || !strings.Contains(*first, "https://first.example/hero.jpg") {
		t.Fatalf("unexpected first markdown %q", *first)
	}

	second := renderMarkdown(ctx, fragment, secondBaseURL)
	if second == nil {
		t.Fatal("expected markdown for second conversion")
	}

	if !strings.Contains(*second, "https://second.example/story") || !strings.Contains(*second, "https://second.example/hero.jpg") {
		t.Fatalf("unexpected second markdown %q", *second)
	}

	if strings.Contains(*second, "https://first.example/") {
		t.Fatalf("expected second markdown to avoid leaked first domain, got %q", *second)
	}

	withoutBase := renderMarkdown(ctx, fragment, nil)
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
