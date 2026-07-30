package memory_test

import (
	"context"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"

	"github.com/MontFerret/contrib/modules/web/html/drivers/memory"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDocumentURLs(t *testing.T) {
	t.Parallel()

	const current = "https://example.com/docs/page.html?lang=en#intro"

	cases := []struct {
		name   string
		markup string
		want   string
	}{
		{
			name:   "missing base",
			markup: `<html><head></head><body></body></html>`,
			want:   current,
		},
		{
			name:   "absolute base",
			markup: `<html><head><base href="https://cdn.example/assets/"></head><body></body></html>`,
			want:   "https://cdn.example/assets/",
		},
		{
			name:   "relative base",
			markup: `<html><head><base href="../assets/"></head><body></body></html>`,
			want:   "https://example.com/assets/",
		},
		{
			name:   "first base wins",
			markup: `<html><head><base href="/first/"><base href="/second/"></head><body></body></html>`,
			want:   "https://example.com/first/",
		},
		{
			name:   "invalid base falls back",
			markup: `<html><head><base href="http://[::1"></head><body></body></html>`,
			want:   current,
		},
		{
			name:   "data base falls back",
			markup: `<html><head><base href="data:text/plain,ignored"></head><body></body></html>`,
			want:   current,
		},
		{
			name:   "javascript base falls back",
			markup: `<html><head><base href="javascript:void(0)"></head><body></body></html>`,
			want:   current,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc := newURLDocument(t, tc.markup, current)
			ctx := context.Background()

			gotCurrent, err := doc.GetCurrentURL(ctx)
			if err != nil {
				t.Fatalf("current URL: %v", err)
			}
			if gotCurrent.String() != current {
				t.Fatalf("expected current URL %q, got %q", current, gotCurrent)
			}

			gotBase, err := doc.GetBaseURL(ctx)
			if err != nil {
				t.Fatalf("base URL: %v", err)
			}
			if gotBase.String() != tc.want {
				t.Fatalf("expected base URL %q, got %q", tc.want, gotBase)
			}
		})
	}
}

func TestDocumentResolveURL(t *testing.T) {
	t.Parallel()

	doc := newURLDocument(
		t,
		`<html><head><base href="/assets/docs/"></head><body></body></html>`,
		"https://example.com/pages/index.html",
	)
	ctx := context.Background()

	cases := []struct {
		name  string
		value string
		want  string
	}{
		{name: "relative path", value: "images/logo.png", want: "https://example.com/assets/docs/images/logo.png"},
		{name: "root relative path", value: "/images/logo.png", want: "https://example.com/images/logo.png"},
		{name: "query", value: "?version=1", want: "https://example.com/assets/docs/?version=1"},
		{name: "fragment", value: "#details", want: "https://example.com/assets/docs/#details"},
		{name: "absolute URL", value: "https://cdn.example/logo.png", want: "https://cdn.example/logo.png"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := doc.ResolveURL(ctx, runtime.NewString(tc.value))
			if err != nil {
				t.Fatalf("resolve URL: %v", err)
			}
			if got.String() != tc.want {
				t.Fatalf("expected resolved URL %q, got %q", tc.want, got)
			}
		})
	}
}

func TestDocumentResolveURLRejectsMalformedInput(t *testing.T) {
	t.Parallel()

	doc := newURLDocument(t, `<html><body></body></html>`, "https://example.com/")

	if _, err := doc.ResolveURL(context.Background(), runtime.NewString("%zz")); err == nil {
		t.Fatal("expected malformed URL error")
	}
}

func newURLDocument(t *testing.T, markup, currentURL string) *memory.HTMLDocument {
	t.Helper()

	parsed, err := goquery.NewDocumentFromReader(strings.NewReader(markup))
	if err != nil {
		t.Fatalf("parse document: %v", err)
	}

	doc, err := memory.NewRootHTMLDocument(parsed, currentURL)
	if err != nil {
		t.Fatalf("create document: %v", err)
	}

	return doc
}
