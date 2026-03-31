package core

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

var (
	positiveKeywords = []string{
		"article", "content", "entry", "post", "story", "main", "body", "text", "prose", "blog", "doc", "docs", "manual", "guide",
	}

	negativeKeywords = []string{
		"comment", "related", "share", "social", "subscribe", "newsletter", "promo", "advert", "ads", "sponsor", "cookie", "sidebar",
		"breadcrumb", "pagination", "pager", "menu", "nav", "footer", "modal", "popup", "recommend", "trending", "rail",
	}
)

type (
	scoredCandidate struct {
		Node       *html.Node
		Score      float64
		Words      int
		TextLength int
	}

	cleanupMetrics struct {
		text           textSummary
		linkTextLength int
		hasKeepContent bool
	}
)

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

func (e *Extractor) cleanCandidate(sel *goquery.Selection, title *string, baseURL *url.URL) extractedBody {
	if sel == nil || sel.Length() == 0 {
		return extractedBody{}
	}

	root := sel.Clone().First()
	removeBoilerplate(root)
	removeDuplicateTitle(root, title)
	removeShortMetadata(root)
	rewriteBodyURLs(root, baseURL)
	pruneEmptyContainers(root)

	textValue := normalizeWhitespace(root.Text())
	if textValue == "" {
		return extractedBody{}
	}

	htmlValue := e.sanitizeHTML(innerHTML(root))
	if htmlValue == nil {
		return extractedBody{}
	}

	return extractedBody{
		Text:      stringPtr(textValue),
		HTML:      htmlValue,
		Excerpt:   firstParagraphExcerpt(root),
		LeadImage: firstImageURL(root, baseURL),
	}
}

func removeBoilerplate(root *goquery.Selection) {
	root.Find("script,style,noscript,nav,aside,footer,form,button,iframe,svg,canvas,template,dialog,menu").Remove()

	if root == nil || len(root.Nodes) == 0 {
		return
	}

	for child := root.Nodes[0].FirstChild; child != nil; {
		next := child.NextSibling
		walkBoilerplateCleanup(child)
		child = next
	}
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
	if root == nil || len(root.Nodes) == 0 {
		return
	}

	for child := root.Nodes[0].FirstChild; child != nil; {
		next := child.NextSibling
		walkEmptyContainerPrune(child)
		child = next
	}
}

func isMeaningfulBody(text string, score float64) bool {
	words := countWords(text)
	if words >= 35 {
		return true
	}

	return len(text) >= 220 && score >= 55
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

func mergeCleanupMetrics(left cleanupMetrics, right cleanupMetrics) cleanupMetrics {
	return cleanupMetrics{
		text:           combineTextSummary(left.text, right.text),
		linkTextLength: left.linkTextLength + right.linkTextLength,
		hasKeepContent: left.hasKeepContent || right.hasKeepContent,
	}
}

func walkBoilerplateCleanup(node *html.Node) cleanupMetrics {
	if node == nil {
		return cleanupMetrics{}
	}

	if node.Type == html.TextNode {
		return cleanupMetrics{text: summarizeTextNode(node.Data)}
	}

	metrics := cleanupMetrics{}
	for child := node.FirstChild; child != nil; {
		next := child.NextSibling
		childMetrics := walkBoilerplateCleanup(child)
		if child.Parent == node {
			metrics = mergeCleanupMetrics(metrics, childMetrics)
		}
		child = next
	}

	if node.Type != html.ElementNode {
		return metrics
	}

	tag := strings.ToLower(node.Data)
	if tag == "a" {
		metrics.linkTextLength += metrics.text.normalizedLength()
	}

	if isKeepContentTag(tag) {
		metrics.hasKeepContent = true
	}

	signals := classIDFromNode(node)
	if signals == "" {
		return metrics
	}

	if hasKeyword(signals, negativeKeywords) && !hasKeyword(signals, positiveKeywords) {
		removeNode(node)

		return cleanupMetrics{}
	}

	if isLinkDensityContainerTag(tag) && measureCleanupLinkDensity(metrics) > 0.72 {
		removeNode(node)

		return cleanupMetrics{}
	}

	return metrics
}

func walkEmptyContainerPrune(node *html.Node) cleanupMetrics {
	if node == nil {
		return cleanupMetrics{}
	}

	if node.Type == html.TextNode {
		return cleanupMetrics{text: summarizeTextNode(node.Data)}
	}

	metrics := cleanupMetrics{}
	for child := node.FirstChild; child != nil; {
		next := child.NextSibling
		childMetrics := walkEmptyContainerPrune(child)
		if child.Parent == node {
			metrics = mergeCleanupMetrics(metrics, childMetrics)
		}
		child = next
	}

	if node.Type != html.ElementNode {
		return metrics
	}

	tag := strings.ToLower(node.Data)
	if isKeepContentTag(tag) {
		metrics.hasKeepContent = true
	}

	if isPrunableContainerTag(tag) && !metrics.hasKeepContent && metrics.text.normalizedLength() == 0 {
		removeNode(node)

		return cleanupMetrics{}
	}

	return metrics
}

func measureCleanupLinkDensity(metrics cleanupMetrics) float64 {
	textLength := metrics.text.normalizedLength()
	if textLength == 0 {
		return 0
	}

	return float64(metrics.linkTextLength) / float64(textLength)
}

func isKeepContentTag(tag string) bool {
	switch tag {
	case "img", "pre", "code", "table", "ul", "ol", "blockquote":
		return true
	default:
		return false
	}
}

func isLinkDensityContainerTag(tag string) bool {
	switch tag {
	case "section", "div", "ul", "ol":
		return true
	default:
		return false
	}
}

func isPrunableContainerTag(tag string) bool {
	switch tag {
	case "div", "section", "span", "header":
		return true
	default:
		return false
	}
}

func removeNode(node *html.Node) {
	if node == nil || node.Parent == nil {
		return
	}

	node.Parent.RemoveChild(node)
}
