package core

import (
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var (
	htmlLangPattern  = regexp.MustCompile(`[a-zA-Z]{2,20}`)
	htmlScopePattern = regexp.MustCompile(`(?i)^(row|col|rowgroup|colgroup)$`)
)

func newHTMLSanitizer() *bluemonday.Policy {
	policy := bluemonday.NewPolicy()

	policy.RequireParseableURLs(true)
	policy.AllowRelativeURLs(true)
	policy.AllowURLSchemes("http", "https", "mailto", "tel")

	policy.AllowElements(
		"a", "abbr", "article", "b", "blockquote", "br", "caption", "cite", "code", "dd", "del", "dfn", "div", "dl", "dt",
		"em", "figcaption", "figure", "h1", "h2", "h3", "h4", "h5", "h6", "header", "hr", "i", "img", "ins", "kbd", "li",
		"mark", "ol", "p", "pre", "q", "s", "samp", "section", "small", "span", "strong", "sub", "sup", "table", "tbody",
		"td", "tfoot", "th", "thead", "time", "tr", "u", "ul", "var", "wbr",
	)

	policy.AllowAttrs("dir").Matching(bluemonday.Direction).Globally()
	policy.AllowAttrs("lang").Matching(htmlLangPattern).Globally()
	policy.AllowAttrs("title").Matching(bluemonday.Paragraph).Globally()

	policy.AllowAttrs("href").OnElements("a")
	policy.AllowAttrs("cite").OnElements("blockquote", "q")
	policy.AllowAttrs("src").OnElements("img")
	policy.AllowAttrs("alt").Matching(bluemonday.Paragraph).OnElements("img")
	policy.AllowAttrs("height", "width").Matching(bluemonday.NumberOrPercent).OnElements("img")
	policy.AllowAttrs("datetime").Matching(bluemonday.ISO8601).OnElements("time")
	policy.AllowAttrs("colspan", "rowspan").Matching(bluemonday.Integer).OnElements("td", "th")
	policy.AllowAttrs("scope").Matching(htmlScopePattern).OnElements("th")

	return policy
}

func (e *Extractor) htmlSanitizerOrNew() *bluemonday.Policy {
	if e != nil && e.htmlSanitizer != nil {
		return e.htmlSanitizer
	}

	return newHTMLSanitizer()
}

func (e *Extractor) sanitizeHTML(fragment string) *string {
	fragment = strings.TrimSpace(fragment)
	if fragment == "" {
		return nil
	}

	sanitized := strings.TrimSpace(e.htmlSanitizerOrNew().Sanitize(fragment))
	if sanitized == "" {
		return nil
	}

	return stringPtr(sanitized)
}
