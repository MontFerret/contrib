package core

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	xhtml "golang.org/x/net/html"
)

type (
	candidateStatsIndex struct {
		stats      map[*xhtml.Node]candidateStats
		candidates []*xhtml.Node
	}

	candidateStats struct {
		text                    textSummary
		linkTextLength          int
		paragraphCount          int
		listItemCount           int
		preCodeCount            int
		tableCount              int
		headingCount            int
		formControlCount        int
		authorSignalCount       int
		bodyStructureCount      int
		negativeCount           int
		hasMatchingTitleHeading bool
	}

	textSummary struct {
		hasRawChars             bool
		startsWithNonWhitespace bool
		endsWithNonWhitespace   bool
		nonWhitespaceBytes      int
		wordCount               int
	}
)

func (summary textSummary) normalizedLength() int {
	if summary.wordCount == 0 {
		return 0
	}

	return summary.nonWhitespaceBytes + summary.wordCount - 1
}

func combineTextSummary(left textSummary, right textSummary) textSummary {
	if !left.hasRawChars {
		return right
	}

	if !right.hasRawChars {
		return left
	}

	wordCount := left.wordCount + right.wordCount
	if left.endsWithNonWhitespace && right.startsWithNonWhitespace && wordCount > 0 {
		wordCount--
	}

	return textSummary{
		hasRawChars:             true,
		startsWithNonWhitespace: left.startsWithNonWhitespace,
		endsWithNonWhitespace:   right.endsWithNonWhitespace,
		nonWhitespaceBytes:      left.nonWhitespaceBytes + right.nonWhitespaceBytes,
		wordCount:               wordCount,
	}
}

func summarizeTextNode(raw string) textSummary {
	if raw == "" {
		return textSummary{}
	}

	summary := textSummary{
		hasRawChars: true,
	}

	inWord := false
	firstRune := true
	lastNonWhitespace := false

	for _, char := range raw {
		isWhitespace := unicode.IsSpace(char)
		if firstRune {
			summary.startsWithNonWhitespace = !isWhitespace
			firstRune = false
		}

		if isWhitespace {
			inWord = false
			lastNonWhitespace = false
			continue
		}

		if !inWord {
			summary.wordCount++
			inWord = true
		}

		summary.nonWhitespaceBytes += utf8.RuneLen(char)
		lastNonWhitespace = true
	}

	summary.endsWithNonWhitespace = lastNonWhitespace

	return summary
}

func mergeCandidateStats(left candidateStats, right candidateStats) candidateStats {
	return candidateStats{
		text:                    combineTextSummary(left.text, right.text),
		linkTextLength:          left.linkTextLength + right.linkTextLength,
		paragraphCount:          left.paragraphCount + right.paragraphCount,
		listItemCount:           left.listItemCount + right.listItemCount,
		preCodeCount:            left.preCodeCount + right.preCodeCount,
		tableCount:              left.tableCount + right.tableCount,
		headingCount:            left.headingCount + right.headingCount,
		formControlCount:        left.formControlCount + right.formControlCount,
		authorSignalCount:       left.authorSignalCount + right.authorSignalCount,
		bodyStructureCount:      left.bodyStructureCount + right.bodyStructureCount,
		negativeCount:           left.negativeCount + right.negativeCount,
		hasMatchingTitleHeading: left.hasMatchingTitleHeading || right.hasMatchingTitleHeading,
	}
}

func buildCandidateStatsIndex(doc *goquery.Document, title string) candidateStatsIndex {
	index := candidateStatsIndex{
		candidates: make([]*xhtml.Node, 0),
		stats:      make(map[*xhtml.Node]candidateStats),
	}

	for _, node := range doc.Nodes {
		walkCandidateStats(node, title, &index)
	}

	return index
}

func walkCandidateStats(node *xhtml.Node, title string, index *candidateStatsIndex) candidateStats {
	switch node.Type {
	case xhtml.TextNode:
		return candidateStats{
			text: summarizeTextNode(node.Data),
		}
	case xhtml.ElementNode:
		if isCandidateNode(node) {
			index.candidates = append(index.candidates, node)
		}
	}

	stats := candidateStats{}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		stats = mergeCandidateStats(stats, walkCandidateStats(child, title, index))
	}

	if node.Type != xhtml.ElementNode {
		return stats
	}

	tag := strings.ToLower(node.Data)
	signals := classIDFromNode(node)

	switch tag {
	case "a":
		stats.linkTextLength += stats.text.normalizedLength()
	case "p":
		stats.paragraphCount++
		stats.bodyStructureCount++
	case "li":
		stats.listItemCount++
	case "pre":
		stats.preCodeCount++
		stats.bodyStructureCount++
	case "code":
		stats.preCodeCount++
	case "table":
		stats.tableCount++
		stats.bodyStructureCount++
	case "h2", "h3", "h4":
		stats.headingCount++
	case "ul", "ol":
		stats.bodyStructureCount++
	}

	if isFormControlTag(tag) {
		stats.formControlCount++
	}

	if isAuthorSignalNode(node, tag) {
		stats.authorSignalCount++
	}

	if hasKeyword(signals, negativeKeywords) {
		stats.negativeCount++
	}

	if title != "" && isTitleHeadingTag(tag) && titleMatches(extractNodeText(node), title) {
		stats.hasMatchingTitleHeading = true
	}

	index.stats[node] = stats

	return stats
}

func scoreCandidate(node *xhtml.Node, stats candidateStats) *scoredCandidate {
	textLength := stats.text.normalizedLength()
	if textLength == 0 {
		return nil
	}

	words := stats.text.wordCount
	if textLength < 80 {
		return nil
	}

	tag := strings.ToLower(node.Data)
	score := tagBonus(tag)

	score += float64(minInt(words, 700)) * 0.18
	score += float64(minInt(stats.paragraphCount, 24)) * 6
	score += float64(minInt(stats.listItemCount, 20)) * 1.5
	score += float64(minInt(stats.preCodeCount, 10)) * 4
	score += float64(minInt(stats.tableCount, 6)) * 5
	score += float64(minInt(stats.headingCount, 8)) * 2

	rootSignals := classIDFromNode(node)
	rootHasNegativeSignal := hasKeyword(rootSignals, negativeKeywords)

	if hasKeyword(rootSignals, positiveKeywords) {
		score += 18
	}

	if rootHasNegativeSignal {
		score -= 28
	}

	score -= measureCachedLinkDensity(stats, textLength) * 95
	score -= float64(stats.formControlCount * 10)

	negativeDescendants := stats.negativeCount
	if rootHasNegativeSignal && negativeDescendants > 0 {
		negativeDescendants--
	}

	score -= float64(minInt(negativeDescendants, 12) * 4)

	if words < 35 {
		score -= 30
	}

	if stats.hasMatchingTitleHeading {
		score += 22
	}

	if stats.authorSignalCount > 0 {
		score += 8
	}

	if stats.bodyStructureCount < 2 {
		score -= 14
	}

	return &scoredCandidate{
		Node:       node,
		Score:      score,
		Words:      words,
		TextLength: textLength,
	}
}

func selectionFromNode(node *xhtml.Node) *goquery.Selection {
	if node == nil {
		return nil
	}

	return goquery.NewDocumentFromNode(node).Selection
}

func measureCachedLinkDensity(stats candidateStats, textLength int) float64 {
	if textLength == 0 {
		return 0
	}

	return float64(stats.linkTextLength) / float64(textLength)
}

func isCandidateNode(node *xhtml.Node) bool {
	if node == nil || node.Type != xhtml.ElementNode {
		return false
	}

	switch strings.ToLower(node.Data) {
	case "article", "main", "section", "div", "body":
		return true
	default:
		return normalizeWhitespace(strings.ToLower(firstAttrFromNode(node, "role"))) == "main"
	}
}

func isFormControlTag(tag string) bool {
	switch tag {
	case "form", "input", "button", "select", "textarea":
		return true
	default:
		return false
	}
}

func isTitleHeadingTag(tag string) bool {
	switch tag {
	case "h1", "h2", "h3":
		return true
	default:
		return false
	}
}

func isAuthorSignalNode(node *xhtml.Node, tag string) bool {
	if tag == "time" {
		return true
	}

	if strings.EqualFold(firstAttrFromNode(node, "rel"), "author") || strings.EqualFold(firstAttrFromNode(node, "itemprop"), "author") {
		return true
	}

	className := strings.ToLower(firstAttrFromNode(node, "class"))

	return hasClassToken(className, "byline") || strings.Contains(className, "author")
}

func extractNodeText(node *xhtml.Node) string {
	if node == nil {
		return ""
	}

	return normalizeWhitespace(goquery.NewDocumentFromNode(node).Text())
}
