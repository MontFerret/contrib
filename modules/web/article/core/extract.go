package core

import (
	"context"
	"math"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"

	"github.com/MontFerret/contrib/modules/web/article/types"
)

var positiveKeywords = []string{
	"article", "content", "entry", "post", "story", "main", "body", "text", "prose", "blog", "doc", "docs", "manual", "guide",
}

var negativeKeywords = []string{
	"comment", "related", "share", "social", "subscribe", "newsletter", "promo", "advert", "ads", "sponsor", "cookie", "sidebar",
	"breadcrumb", "pagination", "pager", "menu", "nav", "footer", "modal", "popup", "recommend", "trending", "rail",
}

type (
	// Source is the normalized article extraction input.
	Source struct {
		SourceURL *url.URL
		TitleHint *string
		HTML      string
	}

	scoredCandidate struct {
		Node       *html.Node
		Score      float64
		Words      int
		TextLength int
	}

	extractedBody struct {
		Text               *string
		HTML               *string
		Markdown           *string
		WordCount          *int
		ReadingTimeMinutes *int
		Excerpt            *string
		LeadImage          *string
	}
)

// Extract returns the best-effort normalized article extracted from raw HTML.
func Extract(ctx context.Context, input string) types.Article {
	return ExtractSource(ctx, Source{HTML: input})
}

// ExtractSource returns the best-effort normalized article extracted from a normalized source.
func ExtractSource(ctx context.Context, source Source) types.Article {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(source.HTML))
	if err != nil {
		return types.Article{}
	}

	baseURL := parseBaseURL(doc)
	if baseURL == nil {
		baseURL = source.SourceURL
	}

	article := extractMetadata(doc, baseURL, source.TitleHint)
	body := extractBody(ctx, doc, article.Title, baseURL)

	article.Text = body.Text
	article.HTML = body.HTML
	article.Markdown = body.Markdown
	article.WordCount = body.WordCount
	article.ReadingTimeMinutes = body.ReadingTimeMinutes

	if article.Excerpt == nil {
		article.Excerpt = body.Excerpt
	}

	if article.LeadImage == nil {
		article.LeadImage = body.LeadImage
	}

	return article
}

func extractBody(ctx context.Context, doc *goquery.Document, title *string, baseURL *url.URL) extractedBody {
	candidate := selectBestCandidate(doc, title)
	if candidate == nil {
		return extractedBody{}
	}

	body := cleanCandidate(selectionFromNode(candidate.Node), title, baseURL)
	if body.Text == nil || !isMeaningfulBody(*body.Text, candidate.Score) {
		return extractedBody{}
	}

	if body.HTML != nil {
		body.Markdown = renderMarkdown(ctx, *body.HTML, baseURL)
	}

	if body.Text != nil {
		wordCount := countWords(*body.Text)
		if wordCount > 0 {
			body.WordCount = intPtr(wordCount)
			body.ReadingTimeMinutes = intPtr(maxInt(1, int(math.Ceil(float64(wordCount)/200.0))))
		}
	}

	return body
}

func selectBestCandidate(doc *goquery.Document, title *string) *scoredCandidate {
	var best *scoredCandidate
	titleValue := valueOrEmpty(title)
	index := buildCandidateStatsIndex(doc, titleValue)

	for _, node := range index.candidates {
		stats, ok := index.stats[node]
		if !ok {
			continue
		}

		candidate := scoreCandidate(node, stats)
		if candidate == nil {
			continue
		}

		if best == nil || candidate.Score > best.Score {
			best = candidate
		}
	}

	return best
}

func tagBonus(tag string) float64 {
	switch tag {
	case "article":
		return 55
	case "main":
		return 34
	case "section":
		return 16
	case "div":
		return 4
	case "body":
		return -22
	default:
		return 0
	}
}

func measureLinkDensity(sel *goquery.Selection, textLength int) float64 {
	if textLength == 0 {
		return 0
	}

	linkLength := 0
	sel.Find("a").Each(func(_ int, link *goquery.Selection) {
		linkLength += len(normalizeWhitespace(link.Text()))
	})

	return float64(linkLength) / float64(textLength)
}

func cleanCandidate(sel *goquery.Selection, title *string, baseURL *url.URL) extractedBody {
	if sel == nil || sel.Length() == 0 {
		return extractedBody{}
	}

	root := sel.Clone().First()
	removeBoilerplate(root)
	removeDuplicateTitle(root, title)
	removeShortMetadata(root)
	rewriteBodyURLs(root, baseURL)
	pruneEmptyContainers(root)

	htmlValue := strings.TrimSpace(innerHTML(root))
	textValue := normalizeWhitespace(root.Text())
	if htmlValue == "" || textValue == "" {
		return extractedBody{}
	}

	return extractedBody{
		Text:      stringPtr(textValue),
		HTML:      stringPtr(htmlValue),
		Excerpt:   firstParagraphExcerpt(root),
		LeadImage: firstImageURL(root, baseURL),
	}
}

func removeBoilerplate(root *goquery.Selection) {
	root.Find("script,style,noscript,nav,aside,footer,form,button,iframe,svg,canvas,template,dialog,menu").Remove()

	root.Find("*").Each(func(_ int, node *goquery.Selection) {
		signals := classID(node)
		if signals == "" {
			return
		}

		if hasKeyword(signals, negativeKeywords) && !hasKeyword(signals, positiveKeywords) {
			node.Remove()
			return
		}

		if node.Is("section,div,ul,ol") && measureLinkDensity(node, len(normalizeWhitespace(node.Text()))) > 0.72 {
			node.Remove()
		}
	})
}

func removeDuplicateTitle(root *goquery.Selection, title *string) {
	if title == nil || *title == "" {
		return
	}

	root.Find("h1,h2").EachWithBreak(func(_ int, node *goquery.Selection) bool {
		text := normalizeWhitespace(node.Text())
		if titleMatches(text, *title) {
			node.Remove()

			return false
		}

		return true
	})
}

func removeShortMetadata(root *goquery.Selection) {
	root.Find("time,[rel='author'],[itemprop='author'],.byline,[class*='byline'],[class*='author'],[class*='date'],[class*='meta'],[id*='author'],[id*='byline'],[id*='date']").Each(func(_ int, node *goquery.Selection) {
		if node.Find("p,pre,table,ul,ol").Length() > 0 {
			return
		}

		text := normalizeWhitespace(node.Text())
		if text != "" && len(text) <= 120 {
			node.Remove()
		}
	})
}

func rewriteBodyURLs(root *goquery.Selection, baseURL *url.URL) {
	if baseURL == nil {
		return
	}

	root.Find("[href],[src]").Each(func(_ int, node *goquery.Selection) {
		if href, ok := node.Attr("href"); ok {
			node.SetAttr("href", resolveURLString(href, baseURL))
		}

		if src, ok := node.Attr("src"); ok {
			node.SetAttr("src", resolveURLString(src, baseURL))
		}
	})
}

func pruneEmptyContainers(root *goquery.Selection) {
	root.Find("div,section,span,header").Each(func(_ int, node *goquery.Selection) {
		if node.Find("img,pre,code,table,ul,ol,blockquote").Length() > 0 {
			return
		}

		text := normalizeWhitespace(node.Text())
		if text == "" {
			node.Remove()
		}
	})
}

func isMeaningfulBody(text string, score float64) bool {
	words := countWords(text)
	if words >= 35 {
		return true
	}

	return len(text) >= 220 && score >= 55
}

func renderMarkdown(ctx context.Context, fragment string, baseURL *url.URL) *string {
	conv := resolveMarkdownConverter(ctx)
	convertOptions := markdownConvertOptions(baseURL)

	markdown, err := conv.ConvertString(fragment, convertOptions...)
	if err != nil {
		return nil
	}

	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return nil
	}

	return stringPtr(markdown)
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}

	return b
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}

	return b
}
