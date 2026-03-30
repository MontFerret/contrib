package core

import (
	"net/url"
	"strings"
	"unicode/utf8"
)

const (
	directiveAllow    = "allow"
	directiveDisallow = "disallow"
)

type candidateRule struct {
	Directive   string
	Pattern     string
	Specificity int
	Order       int
}

// Allows reports whether the path is allowed for the effective user-agent.
func Allows(doc Document, path, userAgent string) bool {
	return Match(doc, path, userAgent).Allowed
}

// Match returns the effective rule-match details for the supplied path.
func Match(doc Document, path, userAgent string) MatchResult {
	agent := normalizeUserAgent(userAgent)
	matchPath := normalizePath(path)

	if isImplicitlyAllowed(matchPath) {
		return MatchResult{
			Allowed:   true,
			UserAgent: agent,
		}
	}

	groups := effectiveGroups(doc.Groups, agent)
	best := chooseBestRule(groups, matchPath)

	if best == nil {
		return MatchResult{
			Allowed:   true,
			UserAgent: agent,
		}
	}

	directive := best.Directive
	pattern := best.Pattern

	return MatchResult{
		Allowed:   best.Directive == directiveAllow,
		Directive: &directive,
		Pattern:   &pattern,
		UserAgent: agent,
	}
}

// SitemapValues returns the declared sitemap URLs.
func SitemapValues(doc Document) []string {
	if len(doc.Sitemaps) == 0 {
		return []string{}
	}

	out := make([]string, len(doc.Sitemaps))
	copy(out, doc.Sitemaps)

	return out
}

func effectiveGroups(groups []Group, userAgent string) []Group {
	exact := make([]Group, 0)
	wildcard := make([]Group, 0)

	for _, group := range groups {
		if groupMatches(group, userAgent) {
			exact = append(exact, group)
			continue
		}

		if groupMatches(group, "*") {
			wildcard = append(wildcard, group)
		}
	}

	if len(exact) > 0 {
		return exact
	}

	return wildcard
}

func groupMatches(group Group, userAgent string) bool {
	for _, candidate := range group.UserAgents {
		if strings.EqualFold(candidate, userAgent) {
			return true
		}
	}

	return false
}

func chooseBestRule(groups []Group, path string) *candidateRule {
	var best *candidateRule
	order := 0

	for _, group := range groups {
		for _, pattern := range group.Allow {
			if matched, specificity := matchPattern(pattern, path); matched {
				best = selectRule(best, candidateRule{
					Directive:   directiveAllow,
					Pattern:     pattern,
					Specificity: specificity,
					Order:       order,
				})
			}

			order++
		}

		for _, pattern := range group.Disallow {
			if matched, specificity := matchPattern(pattern, path); matched {
				best = selectRule(best, candidateRule{
					Directive:   directiveDisallow,
					Pattern:     pattern,
					Specificity: specificity,
					Order:       order,
				})
			}

			order++
		}
	}

	return best
}

func selectRule(current *candidateRule, next candidateRule) *candidateRule {
	if current == nil {
		candidate := next

		return &candidate
	}

	if next.Specificity > current.Specificity {
		candidate := next

		return &candidate
	}

	if next.Specificity < current.Specificity {
		return current
	}

	if current.Directive != next.Directive {
		if next.Directive == directiveAllow {
			candidate := next

			return &candidate
		}

		return current
	}

	if next.Order < current.Order {
		candidate := next

		return &candidate
	}

	return current
}

func matchPattern(pattern, path string) (bool, int) {
	if pattern == "" {
		return false, 0
	}

	normalized := normalizePattern(pattern)
	anchored := strings.HasSuffix(normalized, "$")
	if anchored {
		normalized = normalized[:len(normalized)-1]
	}

	if !matchNormalizedPattern(normalized, path, anchored) {
		return false, 0
	}

	return true, patternSpecificity(normalized)
}

func matchNormalizedPattern(pattern, path string, anchored bool) bool {
	parts := strings.Split(pattern, "*")
	hasWildcard := len(parts) > 1

	// No wildcards: simple prefix or exact match.
	if !hasWildcard {
		if anchored {
			return path == pattern
		}

		return strings.HasPrefix(path, pattern)
	}

	// Pattern is only wildcards (all parts empty): matches everything.
	allEmpty := true
	for _, p := range parts {
		if p != "" {
			allEmpty = false
			break
		}
	}

	if allEmpty {
		return true
	}

	// First literal part must be a prefix of the path.
	if parts[0] != "" {
		if !strings.HasPrefix(path, parts[0]) {
			return false
		}
	}

	// Last literal part, when anchored, must be a suffix of the path.
	if anchored && parts[len(parts)-1] != "" {
		last := parts[len(parts)-1]
		if !strings.HasSuffix(path, last) {
			return false
		}
	}

	// If anchored and pattern ends with *, the trailing wildcard
	// consumes the rest of the path — anchor is satisfied.
	if anchored && parts[len(parts)-1] == "" {
		anchored = false
	}

	// Match interior parts left-to-right.
	offset := 0

	for i, part := range parts {
		if part == "" {
			continue
		}

		if i == 0 {
			// Already verified as prefix above.
			offset = len(part)
			continue
		}

		// For the last part of an anchored pattern, verify the suffix
		// alignment rather than using a greedy first-match.
		if anchored && i == len(parts)-1 {
			suffixStart := len(path) - len(part)
			if suffixStart < offset {
				return false
			}

			return true
		}

		idx := strings.Index(path[offset:], part)
		if idx < 0 {
			return false
		}

		offset += idx + len(part)
	}

	if anchored {
		return offset == len(path)
	}

	return true
}

func patternSpecificity(pattern string) int {
	count := 0

	for i := 0; i < len(pattern); {
		switch pattern[i] {
		case '*':
			i++
		case '%':
			if i+2 < len(pattern) && isHex(pattern[i+1]) && isHex(pattern[i+2]) {
				count++
				i += 3
				continue
			}

			count++
			i++
		default:
			if pattern[i] < utf8.RuneSelf {
				count++
				i++
				continue
			}

			_, size := utf8.DecodeRuneInString(pattern[i:])
			if size <= 0 {
				size = 1
			}

			count += size
			i += size
		}
	}

	return count
}

func normalizeUserAgent(userAgent string) string {
	userAgent = strings.TrimSpace(userAgent)
	if userAgent == "" {
		return "*"
	}

	return userAgent
}

func isImplicitlyAllowed(path string) bool {
	pathOnly := path
	if idx := strings.IndexByte(pathOnly, '?'); idx >= 0 {
		pathOnly = pathOnly[:idx]
	}

	return pathOnly == "/robots.txt"
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/"
	}

	if parsed, err := url.Parse(path); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		path = parsed.EscapedPath()
		if path == "" {
			path = "/"
		}

		if parsed.RawQuery != "" {
			path += "?" + parsed.RawQuery
		}
	}

	if idx := strings.IndexByte(path, '#'); idx >= 0 {
		path = path[:idx]
	}

	if path == "" {
		return "/"
	}

	if path[0] != '/' {
		path = "/" + path
	}

	return normalizeComparable(path)
}

func normalizePattern(pattern string) string {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return ""
	}

	if pattern[0] != '/' {
		pattern = "/" + pattern
	}

	return normalizeComparable(pattern)
}

func normalizeComparable(input string) string {
	var b strings.Builder
	b.Grow(len(input))

	for i := 0; i < len(input); {
		if input[i] == '%' && i+2 < len(input) && isHex(input[i+1]) && isHex(input[i+2]) {
			octet := decodeHexByte(input[i+1], input[i+2])
			if isUnreserved(octet) {
				b.WriteByte(octet)
			} else {
				writeEscape(&b, octet)
			}

			i += 3
			continue
		}

		if input[i] < utf8.RuneSelf {
			ch := input[i]
			if ch < 0x20 || ch == 0x7f || ch == ' ' {
				writeEscape(&b, ch)
			} else {
				b.WriteByte(ch)
			}

			i++
			continue
		}

		r, size := utf8.DecodeRuneInString(input[i:])
		if r == utf8.RuneError && size == 1 {
			writeEscape(&b, input[i])
			i++
			continue
		}

		var buf [utf8.UTFMax]byte
		n := utf8.EncodeRune(buf[:], r)
		for _, ch := range buf[:n] {
			writeEscape(&b, ch)
		}

		i += size
	}

	return b.String()
}

func writeEscape(b *strings.Builder, value byte) {
	const hex = "0123456789ABCDEF"

	b.WriteByte('%')
	b.WriteByte(hex[value>>4])
	b.WriteByte(hex[value&0x0f])
}

func decodeHexByte(a, c byte) byte {
	return hexNibble(a)<<4 | hexNibble(c)
}

func hexNibble(ch byte) byte {
	switch {
	case ch >= '0' && ch <= '9':
		return ch - '0'
	case ch >= 'a' && ch <= 'f':
		return ch - 'a' + 10
	default:
		return ch - 'A' + 10
	}
}

func isHex(ch byte) bool {
	return (ch >= '0' && ch <= '9') ||
		(ch >= 'a' && ch <= 'f') ||
		(ch >= 'A' && ch <= 'F')
}

func isUnreserved(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '-' ||
		ch == '.' ||
		ch == '_' ||
		ch == '~'
}
