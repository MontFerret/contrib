package core

import (
	"net/url"

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

type scoredCandidate struct {
	Node       *html.Node
	Score      float64
	Words      int
	TextLength int
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
