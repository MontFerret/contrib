package core

import (
	"bytes"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	xhtml "golang.org/x/net/html"
)

var whitespacePattern = regexp.MustCompile(`\s+`)

var timestampLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	time.RFC1123Z,
	time.RFC1123,
	time.RFC822Z,
	time.RFC822,
	time.RFC850,
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	"2006-01-02",
	"2006-01-02 15:04",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04:05 -0700",
	"2006-01-02T15:04:05-0700",
	"2006-01-02T15:04:05",
	"Mon, 02 Jan 2006 15:04:05 GMT",
	"Mon, 02 Jan 2006 15:04:05 MST",
	"January 2, 2006 3:04 PM MST",
	"January 2, 2006 3:04 PM",
	"January 2, 2006",
	"Jan 2, 2006 3:04 PM MST",
	"Jan 2, 2006 3:04 PM",
	"Jan 2, 2006",
}

func normalizeWhitespace(input string) string {
	return strings.TrimSpace(whitespacePattern.ReplaceAllString(input, " "))
}

func comparableText(input string) string {
	normalized := strings.ToLower(normalizeWhitespace(input))
	if normalized == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(normalized))

	for _, char := range normalized {
		switch {
		case unicode.IsLetter(char), unicode.IsNumber(char), unicode.IsSpace(char):
			builder.WriteRune(char)
		default:
			builder.WriteByte(' ')
		}
	}

	return normalizeWhitespace(builder.String())
}

func titleMatches(input string, title string) bool {
	left := comparableText(input)
	right := comparableText(title)
	if left == "" || right == "" {
		return false
	}

	if left == right {
		return true
	}

	if strings.HasPrefix(left, right) || strings.HasPrefix(right, left) {
		return true
	}

	return strings.Contains(left, right) && len(right) >= 24
}

func stringPtr(value string) *string {
	value = normalizeWhitespace(value)
	if value == "" {
		return nil
	}

	out := value

	return &out
}

func intPtr(value int) *int {
	out := value

	return &out
}

func firstAttr(sel *goquery.Selection, name string) string {
	if sel == nil || sel.Length() == 0 {
		return ""
	}

	return normalizeWhitespace(sel.AttrOr(name, ""))
}

func parseBaseURL(doc *goquery.Document) *url.URL {
	raw := firstAttr(doc.Find("base[href]").First(), "href")
	if raw == "" {
		return nil
	}

	parsed, err := url.Parse(raw)
	if err != nil || !parsed.IsAbs() {
		return nil
	}

	return parsed
}

func resolveURLString(raw string, baseURL *url.URL) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	if parsed.IsAbs() || baseURL == nil {
		return raw
	}

	return baseURL.ResolveReference(parsed).String()
}

func resolveURLPtr(raw string, baseURL *url.URL) *string {
	resolved := resolveURLString(raw, baseURL)
	if resolved == "" {
		return nil
	}

	return &resolved
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))

	for _, value := range values {
		normalized := normalizeWhitespace(value)
		if normalized == "" {
			continue
		}

		key := strings.ToLower(normalized)
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		result = append(result, normalized)
	}

	return result
}

func parseDelimitedValues(values ...string) []string {
	result := make([]string, 0)

	for _, value := range values {
		value = normalizeWhitespace(value)
		if value == "" {
			continue
		}

		parts := strings.FieldsFunc(value, func(char rune) bool {
			return char == ',' || char == ';' || char == '|'
		})

		if len(parts) == 0 {
			result = append(result, value)
			continue
		}

		for _, part := range parts {
			part = normalizeWhitespace(part)
			if part != "" {
				result = append(result, part)
			}
		}
	}

	return uniqueStrings(result)
}

func normalizeTimestamp(raw string) *string {
	raw = normalizeWhitespace(raw)
	if raw == "" {
		return nil
	}

	for _, layout := range timestampLayouts {
		parsed, err := time.Parse(layout, raw)
		if err == nil {
			out := parsed.UTC().Format(time.RFC3339)

			return &out
		}
	}

	out := raw

	return &out
}

func countWords(input string) int {
	return len(strings.Fields(input))
}

func truncateText(input string, limit int) string {
	input = normalizeWhitespace(input)
	if input == "" || len(input) <= limit {
		return input
	}

	cut := strings.LastIndexByte(input[:limit], ' ')
	if cut <= 0 {
		cut = limit
	}

	return strings.TrimSpace(input[:cut]) + "..."
}

func innerHTML(sel *goquery.Selection) string {
	if sel == nil || len(sel.Nodes) == 0 {
		return ""
	}

	var buffer bytes.Buffer
	for child := sel.Nodes[0].FirstChild; child != nil; child = child.NextSibling {
		if err := xhtml.Render(&buffer, child); err != nil {
			return ""
		}
	}

	return strings.TrimSpace(buffer.String())
}

func classID(sel *goquery.Selection) string {
	return strings.ToLower(strings.Join([]string{
		firstAttr(sel, "id"),
		firstAttr(sel, "class"),
		firstAttr(sel, "role"),
		firstAttr(sel, "itemprop"),
	}, " "))
}

func hasKeyword(input string, keywords []string) bool {
	input = strings.ToLower(input)
	for _, keyword := range keywords {
		if strings.Contains(input, keyword) {
			return true
		}
	}

	return false
}

func cleanByline(value string) *string {
	value = normalizeWhitespace(value)
	if value == "" {
		return nil
	}

	lowered := strings.ToLower(value)
	if strings.HasPrefix(lowered, "by ") && len(value) > 3 {
		value = normalizeWhitespace(value[3:])
	}

	return stringPtr(value)
}

func firstTextBySelectors(root *goquery.Selection, selectors []string, minLength int, maxLength int) *string {
	for _, selector := range selectors {
		var found *string

		root.Find(selector).EachWithBreak(func(_ int, sel *goquery.Selection) bool {
			text := normalizeWhitespace(sel.Text())
			if text == "" {
				return true
			}

			if minLength > 0 && len(text) < minLength {
				return true
			}

			if maxLength > 0 && len(text) > maxLength {
				return true
			}

			found = stringPtr(text)

			return false
		})

		if found != nil {
			return found
		}
	}

	return nil
}

func firstImageURL(root *goquery.Selection, baseURL *url.URL) *string {
	if root == nil || root.Length() == 0 {
		return nil
	}

	var found *string

	root.Find("img[src]").EachWithBreak(func(_ int, sel *goquery.Selection) bool {
		src := firstAttr(sel, "src")
		if src == "" {
			return true
		}

		found = resolveURLPtr(src, baseURL)

		return false
	})

	return found
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
