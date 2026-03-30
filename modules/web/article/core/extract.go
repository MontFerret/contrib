package core

import (
	"math"
	"net/url"
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
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
	scoredCandidate struct {
		Selection  *goquery.Selection
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
func Extract(input string) types.Article {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(input))
	if err != nil {
		return types.Article{}
	}

	baseURL := parseBaseURL(doc)
	article := extractMetadata(doc, baseURL)
	body := extractBody(doc, article.Title, baseURL)

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

func extractBody(doc *goquery.Document, title *string, baseURL *url.URL) extractedBody {
	candidate := selectBestCandidate(doc, title)
	if candidate == nil {
		return extractedBody{}
	}

	body := cleanCandidate(candidate.Selection, title, baseURL)
	if body.Text == nil || !isMeaningfulBody(*body.Text, candidate.Score) {
		return extractedBody{}
	}

	if body.HTML != nil {
		body.Markdown = renderMarkdown(*body.HTML, baseURL)
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
	seen := make(map[*html.Node]struct{})
	titleValue := valueOrEmpty(title)

	doc.Find("article,main,[role='main'],section,div,body").Each(func(_ int, sel *goquery.Selection) {
		if len(sel.Nodes) == 0 {
			return
		}

		if _, ok := seen[sel.Nodes[0]]; ok {
			return
		}

		seen[sel.Nodes[0]] = struct{}{}

		candidate := scoreCandidate(sel, titleValue)
		if candidate == nil {
			return
		}

		if best == nil || candidate.Score > best.Score {
			best = candidate
		}
	})

	return best
}

func scoreCandidate(sel *goquery.Selection, title string) *scoredCandidate {
	text := normalizeWhitespace(sel.Text())
	if text == "" {
		return nil
	}

	words := countWords(text)
	textLength := len(text)
	if textLength < 80 {
		return nil
	}

	tag := strings.ToLower(goquery.NodeName(sel))
	score := tagBonus(tag)

	score += float64(minInt(words, 700)) * 0.18
	score += float64(minInt(sel.Find("p").Length(), 24)) * 6
	score += float64(minInt(sel.Find("li").Length(), 20)) * 1.5
	score += float64(minInt(sel.Find("pre, code").Length(), 10)) * 4
	score += float64(minInt(sel.Find("table").Length(), 6)) * 5
	score += float64(minInt(sel.Find("h2, h3, h4").Length(), 8)) * 2

	rootSignals := classID(sel)
	if hasKeyword(rootSignals, positiveKeywords) {
		score += 18
	}

	if hasKeyword(rootSignals, negativeKeywords) {
		score -= 28
	}

	linkDensity := measureLinkDensity(sel, textLength)
	score -= linkDensity * 95

	formPenalty := sel.Find("form,input,button,select,textarea").Length()
	score -= float64(formPenalty * 10)

	negativeDescendants := countNegativeDescendants(sel)
	score -= float64(minInt(negativeDescendants, 12) * 4)

	if words < 35 {
		score -= 30
	}

	if title != "" && selectionMentionsTitle(sel, title) {
		score += 22
	}

	if sel.Find("time,[rel='author'],[itemprop='author'],.byline,[class*='author']").Length() > 0 {
		score += 8
	}

	if sel.Find("p,pre,table,ul,ol").Length() < 2 {
		score -= 14
	}

	return &scoredCandidate{
		Selection:  sel,
		Score:      score,
		Words:      words,
		TextLength: textLength,
	}
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

func countNegativeDescendants(sel *goquery.Selection) int {
	count := 0

	sel.Find("*").Each(func(_ int, node *goquery.Selection) {
		if hasKeyword(classID(node), negativeKeywords) {
			count++
		}
	})

	return count
}

func selectionMentionsTitle(sel *goquery.Selection, title string) bool {
	found := false

	sel.Find("h1,h2,h3").EachWithBreak(func(_ int, node *goquery.Selection) bool {
		if titleMatches(node.Text(), title) {
			found = true

			return false
		}

		return true
	})

	return found
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

func renderMarkdown(fragment string, baseURL *url.URL) *string {
	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			table.NewTablePlugin(
				table.WithHeaderPromotion(true),
			),
		),
	)

	convertOptions := make([]converter.ConvertOptionFunc, 0, 1)
	if baseURL != nil {
		convertOptions = append(convertOptions, converter.WithDomain(baseURL.String()))
	}

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
