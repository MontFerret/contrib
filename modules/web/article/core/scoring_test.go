package core

import (
	"strconv"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestBuildCandidateStatsIndexMatchesLiveSelectionMetrics(t *testing.T) {
	doc := mustArticleDocument(t, `
		<html>
		  <body>
		    <main id="content" class="docs-content">
		      <h2><span>Deep</span><span>Feature</span></h2>
		      <p><span>Hel</span><span>lo</span> world for article extraction stats.</p>
		      <p class="byline">By Jane Doe</p>
		      <div class="related-links">
		        <a href="/share">Share story</a>
		      </div>
		      <ul>
		        <li>First</li>
		        <li>Second</li>
		      </ul>
		      <pre><code>curl https://example.com</code></pre>
		      <table><tr><td>Value</td></tr></table>
		      <time datetime="2026-03-31">March 31, 2026</time>
		    </main>
		  </body>
		</html>
	`)

	index := buildCandidateStatsIndex(doc, "DeepFeature")
	main := doc.Find("main").First()
	if main.Length() == 0 {
		t.Fatal("expected main candidate")
	}

	stats, ok := index.stats[main.Nodes[0]]
	if !ok {
		t.Fatal("expected stats for main candidate")
	}

	liveText := normalizeWhitespace(main.Text())
	if got := stats.text.normalizedLength(); got != len(liveText) {
		t.Fatalf("unexpected text length %d, want %d", got, len(liveText))
	}

	if got := stats.text.wordCount; got != countWords(liveText) {
		t.Fatalf("unexpected word count %d, want %d", got, countWords(liveText))
	}

	liveLinkLength := 0
	main.Find("a").Each(func(_ int, link *goquery.Selection) {
		liveLinkLength += len(normalizeWhitespace(link.Text()))
	})
	if stats.linkTextLength != liveLinkLength {
		t.Fatalf("unexpected link text length %d, want %d", stats.linkTextLength, liveLinkLength)
	}

	liveNegativeCount := 0
	main.Find("*").Each(func(_ int, node *goquery.Selection) {
		if hasKeyword(classID(node), negativeKeywords) {
			liveNegativeCount++
		}
	})
	if stats.negativeCount != liveNegativeCount {
		t.Fatalf("unexpected negative count %d, want %d", stats.negativeCount, liveNegativeCount)
	}

	if stats.paragraphCount != main.Find("p").Length() {
		t.Fatalf("unexpected paragraph count %d, want %d", stats.paragraphCount, main.Find("p").Length())
	}

	if stats.listItemCount != main.Find("li").Length() {
		t.Fatalf("unexpected list item count %d, want %d", stats.listItemCount, main.Find("li").Length())
	}

	if stats.preCodeCount != main.Find("pre, code").Length() {
		t.Fatalf("unexpected pre/code count %d, want %d", stats.preCodeCount, main.Find("pre, code").Length())
	}

	if stats.tableCount != main.Find("table").Length() {
		t.Fatalf("unexpected table count %d, want %d", stats.tableCount, main.Find("table").Length())
	}

	if stats.headingCount != main.Find("h2, h3, h4").Length() {
		t.Fatalf("unexpected heading count %d, want %d", stats.headingCount, main.Find("h2, h3, h4").Length())
	}

	if stats.authorSignalCount != main.Find("time,[rel='author'],[itemprop='author'],.byline,[class*='author']").Length() {
		t.Fatalf(
			"unexpected author signal count %d, want %d",
			stats.authorSignalCount,
			main.Find("time,[rel='author'],[itemprop='author'],.byline,[class*='author']").Length(),
		)
	}

	if stats.bodyStructureCount != main.Find("p,pre,table,ul,ol").Length() {
		t.Fatalf(
			"unexpected body structure count %d, want %d",
			stats.bodyStructureCount,
			main.Find("p,pre,table,ul,ol").Length(),
		)
	}

	if !stats.hasMatchingTitleHeading {
		t.Fatal("expected matching title heading signal")
	}
}

func TestExtractLargeNestedPageStillFindsArticle(t *testing.T) {
	article := NewExtractor().Extract(buildLargeCandidateFixture(120))

	if article.Title == nil || *article.Title != "Deep Feature" {
		t.Fatalf("unexpected title %+v", article.Title)
	}

	if article.SiteName == nil || *article.SiteName != "Example Daily" {
		t.Fatalf("unexpected siteName %+v", article.SiteName)
	}

	if article.Text == nil || !strings.Contains(*article.Text, "single-pass candidate scoring cache") {
		t.Fatalf("unexpected text %+v", article.Text)
	}

	if strings.Contains(*article.Text, "Newsletter signup") || strings.Contains(*article.Text, "Related story") {
		t.Fatalf("expected article body to exclude repeated noise, got %q", *article.Text)
	}
}

func mustArticleDocument(t *testing.T, input string) *goquery.Document {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	return doc
}

func buildLargeCandidateFixture(noiseBlocks int) string {
	var builder strings.Builder

	builder.WriteString(`<html><head><title>Deep Feature | Example Daily</title></head><body>`)
	for i := 0; i < noiseBlocks; i++ {
		builder.WriteString(`<section class="related rail">`)
		builder.WriteString(`<div class="newsletter card">Newsletter signup for daily alerts.</div>`)
		for j := 0; j < 4; j++ {
			builder.WriteString(`<div class="related-story-card"><a href="/related-`)
			builder.WriteString(strconv.Itoa(i))
			builder.WriteString(`-`)
			builder.WriteString(strconv.Itoa(j))
			builder.WriteString(`">Related story `)
			builder.WriteString(strconv.Itoa(i))
			builder.WriteString(`.`)
			builder.WriteString(strconv.Itoa(j))
			builder.WriteString(` with more links and promo text.</a></div>`)
		}
		builder.WriteString(`</section>`)
	}

	builder.WriteString(`<div class="page-shell">`)
	for i := 0; i < 40; i++ {
		builder.WriteString(`<div class="content-shell">`)
	}

	builder.WriteString(`
		<article class="story article-content">
		  <h1>Deep Feature</h1>
		  <p>The single-pass candidate scoring cache should still identify the primary article body even when the page contains a large number of nested div and section containers competing for attention.</p>
		  <p>This paragraph keeps the body comfortably above the meaningful-text threshold while the surrounding rail, promo, and related-story blocks try to pull the heuristics away from the actual article content.</p>
		</article>
	`)

	for i := 0; i < 40; i++ {
		builder.WriteString(`</div>`)
	}
	builder.WriteString(`</div></body></html>`)

	return builder.String()
}
