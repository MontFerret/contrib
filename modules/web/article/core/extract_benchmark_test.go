package core

import (
	"context"
	"fmt"
	"net/url"
	"testing"
)

func BenchmarkExtractLargeCandidatePage(b *testing.B) {
	fixture := buildLargeCandidateFixture(180)
	ctx := WithExtractor(context.Background(), NewExtractor())
	extractor := ExtractorFromContext(ctx)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		article := extractor.Extract(fixture)
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

	ctx := WithExtractor(context.Background(), NewExtractor())
	extractor := ExtractorFromContext(ctx)

	b.Run("no_base", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			markdown := extractor.renderMarkdown(fragment, nil)
			if markdown == nil {
				b.Fatal("expected markdown without base")
			}
		}
	})

	b.Run("with_base", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			markdown := extractor.renderMarkdown(fragment, baseURL)
			if markdown == nil {
				b.Fatal("expected markdown with base")
			}
		}
	})
}

func BenchmarkExtractCleanupScaling(b *testing.B) {
	ctx := WithExtractor(context.Background(), NewExtractor())
	extractor := ExtractorFromContext(ctx)

	for _, blocks := range []int{120, 240, 480} {
		fixture := buildLargeCandidateFixture(blocks)

		b.Run(fmt.Sprintf("noise_%d", blocks), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				article := extractor.Extract(fixture)
				if article.Text == nil {
					b.Fatal("expected article text")
				}
			}
		})
	}
}
