package core

import (
	"math"
	"net/url"
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"

	"github.com/MontFerret/contrib/modules/web/article/types"
)

type (
	// Source is the normalized article extraction input.
	Source struct {
		SourceURL *url.URL
		TitleHint *string
		HTML      string
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

	Extractor struct {
		markdownConverter *converter.Converter
		htmlSanitizer     *bluemonday.Policy
	}
)

func NewExtractor() *Extractor {
	return &Extractor{
		markdownConverter: newMarkdownConverter(),
		htmlSanitizer:     newHTMLSanitizer(),
	}
}

// Extract returns the best-effort normalized article extracted from raw HTML.
func (e *Extractor) Extract(input string) types.Article {
	return e.ExtractSource(Source{HTML: input})
}

// ExtractSource returns the best-effort normalized article extracted from a normalized source.
func (e *Extractor) ExtractSource(source Source) types.Article {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(source.HTML))
	if err != nil {
		return types.Article{}
	}

	baseURL := parseBaseURL(doc, source.SourceURL)
	if baseURL == nil {
		baseURL = source.SourceURL
	}

	article := extractMetadata(doc, baseURL, source.TitleHint)
	body := e.extractBody(doc, article.Title, baseURL)

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

func (e *Extractor) extractBody(doc *goquery.Document, title *string, baseURL *url.URL) extractedBody {
	candidate := selectBestCandidate(doc, title)
	if candidate == nil {
		return extractedBody{}
	}

	body := e.cleanCandidate(selectionFromNode(candidate.Node), title, baseURL)
	if body.Text == nil || !isMeaningfulBody(*body.Text, candidate.Score) {
		return extractedBody{}
	}

	if body.HTML != nil {
		body.Markdown = e.renderMarkdown(*body.HTML, baseURL)
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

func (e *Extractor) renderMarkdown(fragment string, baseURL *url.URL) *string {
	conv := e.markdownConverterOrNew()
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

func (e *Extractor) markdownConverterOrNew() *converter.Converter {
	if e != nil && e.markdownConverter != nil {
		return e.markdownConverter
	}

	return newMarkdownConverter()
}
