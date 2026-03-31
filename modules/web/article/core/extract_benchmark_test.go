package core

import (
	"context"
	"net/url"
	"testing"
)

func BenchmarkExtractLargeCandidatePage(b *testing.B) {
	fixture := buildLargeCandidateFixture(180)
	ctx := WithMarkdownConverter(context.Background(), NewMarkdownConverter())

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		article := Extract(ctx, fixture)
		if article.Text == nil {
			b.Fatal("expected article text")
		}
	}
}

func BenchmarkRenderMarkdown(b *testing.B) {
	fragment := `
		<p>Markdown conversion should reuse the initialized plugin set instead of rebuilding it for every extraction.</p>
		<p><a href="/story">Story</a> <img src="/hero.jpg" alt="Lead" /></p>
		<table>
		  <tr><td>Field</td><td>Value</td></tr>
		  <tr><td>kind</td><td>benchmark</td></tr>
		</table>
	`

	baseURL, err := url.Parse("https://example.com/posts/alpha")
	if err != nil {
		b.Fatalf("unexpected base url parse error: %v", err)
	}

	ctx := WithMarkdownConverter(context.Background(), NewMarkdownConverter())

	b.Run("no_base", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			markdown := renderMarkdown(ctx, fragment, nil)
			if markdown == nil {
				b.Fatal("expected markdown without base")
			}
		}
	})

	b.Run("with_base", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			markdown := renderMarkdown(ctx, fragment, baseURL)
			if markdown == nil {
				b.Fatal("expected markdown with base")
			}
		}
	})
}
