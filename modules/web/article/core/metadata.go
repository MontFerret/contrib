package core

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/MontFerret/contrib/modules/web/article/types"
)

var articleTypes = map[string]struct{}{
	"article":             {},
	"analysisnewsarticle": {},
	"blogposting":         {},
	"newsarticle":         {},
	"reportage":           {},
	"scholarlyarticle":    {},
	"techarticle":         {},
}

type metaIndex map[string][]string

func extractMetadata(doc *goquery.Document, baseURL *url.URL, titleHint *string) types.Article {
	article := types.Article{}

	mergeArticle(&article, extractJSONLDMetadata(doc, baseURL))
	mergeArticle(&article, extractMetaTagMetadata(doc, baseURL))
	mergeArticle(&article, extractDOMFallbackMetadata(doc, baseURL))
	fillTitleAndSiteFromText(doc.Find("title").First().Text(), &article)
	fillTitleAndSiteFromText(valueOrEmpty(titleHint), &article)

	return article
}

func mergeArticle(dst *types.Article, src types.Article) {
	if dst.Title == nil {
		dst.Title = src.Title
	}

	if dst.Byline == nil {
		dst.Byline = src.Byline
	}

	if dst.Excerpt == nil {
		dst.Excerpt = src.Excerpt
	}

	if dst.SiteName == nil {
		dst.SiteName = src.SiteName
	}

	if dst.PublishedAt == nil {
		dst.PublishedAt = src.PublishedAt
	}

	if dst.UpdatedAt == nil {
		dst.UpdatedAt = src.UpdatedAt
	}

	if dst.Lang == nil {
		dst.Lang = src.Lang
	}

	if dst.Dir == nil {
		dst.Dir = src.Dir
	}

	if dst.CanonicalURL == nil {
		dst.CanonicalURL = src.CanonicalURL
	}

	if dst.LeadImage == nil {
		dst.LeadImage = src.LeadImage
	}

	if dst.Text == nil {
		dst.Text = src.Text
	}

	if dst.HTML == nil {
		dst.HTML = src.HTML
	}

	if dst.Markdown == nil {
		dst.Markdown = src.Markdown
	}

	if dst.WordCount == nil {
		dst.WordCount = src.WordCount
	}

	if dst.ReadingTimeMinutes == nil {
		dst.ReadingTimeMinutes = src.ReadingTimeMinutes
	}

	if len(dst.Tags) == 0 && len(src.Tags) > 0 {
		dst.Tags = append([]string(nil), src.Tags...)
	}

	if len(dst.Categories) == 0 && len(src.Categories) > 0 {
		dst.Categories = append([]string(nil), src.Categories...)
	}
}

func extractJSONLDMetadata(doc *goquery.Document, baseURL *url.URL) types.Article {
	article := types.Article{}

	doc.Find("script[type='application/ld+json']").Each(func(_ int, sel *goquery.Selection) {
		payload := strings.TrimSpace(sel.Text())
		if payload == "" {
			return
		}

		var decoded any
		if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
			return
		}

		for _, node := range collectArticleNodes(decoded) {
			mergeArticle(&article, articleFromJSONLD(node, baseURL))
		}
	})

	return article
}

func collectArticleNodes(value any) []map[string]any {
	result := make([]map[string]any, 0)

	var walk func(any)
	walk = func(node any) {
		switch typed := node.(type) {
		case []any:
			for _, item := range typed {
				walk(item)
			}
		case map[string]any:
			if isArticleNode(typed) {
				result = append(result, typed)
			}

			for _, item := range typed {
				walk(item)
			}
		}
	}

	walk(value)

	return result
}

func isArticleNode(node map[string]any) bool {
	typeValue, ok := node["@type"]
	if !ok {
		return false
	}

	switch typed := typeValue.(type) {
	case string:
		_, ok = articleTypes[strings.ToLower(typed)]

		return ok
	case []any:
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				continue
			}

			if _, found := articleTypes[strings.ToLower(text)]; found {
				return true
			}
		}
	}

	return false
}

func articleFromJSONLD(node map[string]any, baseURL *url.URL) types.Article {
	return types.Article{
		Title:       extractJSONLDString(node["headline"], node["name"]),
		Byline:      cleanByline(firstJSONLDName(node["author"])),
		Excerpt:     extractJSONLDString(node["description"]),
		SiteName:    extractJSONLDPublisherName(node["publisher"]),
		PublishedAt: normalizeTimestamp(firstJSONLDString(node["datePublished"], node["dateCreated"])),
		UpdatedAt:   normalizeTimestamp(firstJSONLDString(node["dateModified"])),
		Lang:        extractJSONLDString(node["inLanguage"]),
		CanonicalURL: resolveURLPtr(
			firstJSONLDString(node["url"], node["mainEntityOfPage"], node["@id"]),
			baseURL,
		),
		LeadImage:  extractJSONLDURL(node["image"], baseURL),
		Tags:       parseDelimitedValues(firstJSONLDStrings(node["keywords"])...),
		Categories: uniqueStrings(firstJSONLDStrings(node["articleSection"])),
	}
}

func extractJSONLDPublisherName(value any) *string {
	switch typed := value.(type) {
	case string:
		return stringPtr(typed)
	case []any:
		for _, item := range typed {
			if name := extractJSONLDPublisherName(item); name != nil {
				return name
			}
		}
	case map[string]any:
		return extractJSONLDString(typed["name"])
	}

	return nil
}

func firstJSONLDName(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []any:
		for _, item := range typed {
			if name := firstJSONLDName(item); name != "" {
				return name
			}
		}
	case map[string]any:
		return firstJSONLDString(typed["name"], typed["alternateName"])
	}

	return ""
}

func extractJSONLDURL(value any, baseURL *url.URL) *string {
	switch typed := value.(type) {
	case string:
		return resolveURLPtr(typed, baseURL)
	case []any:
		for _, item := range typed {
			if found := extractJSONLDURL(item, baseURL); found != nil {
				return found
			}
		}
	case map[string]any:
		return resolveURLPtr(firstJSONLDString(typed["url"], typed["contentUrl"], typed["@id"]), baseURL)
	}

	return nil
}

func extractJSONLDString(values ...any) *string {
	return stringPtr(firstJSONLDString(values...))
}

func firstJSONLDString(values ...any) string {
	for _, value := range values {
		switch typed := value.(type) {
		case string:
			if normalizeWhitespace(typed) != "" {
				return typed
			}
		case map[string]any:
			if text := firstJSONLDString(typed["url"], typed["@id"], typed["name"]); text != "" {
				return text
			}
		}
	}

	return ""
}

func firstJSONLDStrings(value any) []string {
	switch typed := value.(type) {
	case string:
		return []string{typed}
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := firstJSONLDString(item); text != "" {
				result = append(result, text)
			}
		}

		return uniqueStrings(result)
	default:
		return nil
	}
}

func extractMetaTagMetadata(doc *goquery.Document, baseURL *url.URL) types.Article {
	index := buildMetaIndex(doc)

	return types.Article{
		Title:       stringPtr(firstMetaValue(index, "og:title", "twitter:title", "parsely-title", "title")),
		Byline:      cleanByline(firstMetaValue(index, "author", "article:author", "parsely-author")),
		Excerpt:     stringPtr(firstMetaValue(index, "description", "og:description", "twitter:description", "parsely-description")),
		SiteName:    stringPtr(firstMetaValue(index, "og:site_name", "application-name")),
		PublishedAt: normalizeTimestamp(firstMetaValue(index, "article:published_time", "datepublished", "pubdate", "parsely-pub-date")),
		UpdatedAt: normalizeTimestamp(
			firstMetaValue(index, "article:modified_time", "datemodified", "lastmod", "parsely-modified-date"),
		),
		Lang:         stringPtr(firstMetaValue(index, "content-language", "language")),
		CanonicalURL: resolveURLPtr(firstMetaValue(index, "og:url"), baseURL),
		LeadImage: resolveURLPtr(
			firstMetaValue(index, "og:image", "og:image:url", "twitter:image", "image"),
			baseURL,
		),
		Tags: uniqueStrings(append(
			parseDelimitedValues(firstMetaValue(index, "keywords", "news_keywords")),
			allMetaValues(index, "article:tag")...,
		)),
		Categories: uniqueStrings(allMetaValues(index, "article:section", "section")),
	}
}

func buildMetaIndex(doc *goquery.Document) metaIndex {
	index := make(metaIndex)

	doc.Find("meta").Each(func(_ int, sel *goquery.Selection) {
		key := ""
		for _, attr := range []string{"property", "name", "itemprop"} {
			value := strings.ToLower(firstAttr(sel, attr))
			if value != "" {
				key = value
				break
			}
		}

		content := firstAttr(sel, "content")
		if key == "" || content == "" {
			return
		}

		index[key] = append(index[key], content)
	})

	return index
}

func firstMetaValue(index metaIndex, keys ...string) string {
	for _, key := range keys {
		values := index[strings.ToLower(key)]
		for _, value := range values {
			value = normalizeWhitespace(value)
			if value != "" {
				return value
			}
		}
	}

	return ""
}

func allMetaValues(index metaIndex, keys ...string) []string {
	result := make([]string, 0)

	for _, key := range keys {
		values := index[strings.ToLower(key)]
		for _, value := range values {
			value = normalizeWhitespace(value)
			if value != "" {
				result = append(result, value)
			}
		}
	}

	return uniqueStrings(result)
}

func extractDOMFallbackMetadata(doc *goquery.Document, baseURL *url.URL) types.Article {
	return types.Article{
		Lang:         stringPtr(firstAttr(doc.Find("html").First(), "lang")),
		Dir:          stringPtr(firstAttr(doc.Find("html").First(), "dir")),
		CanonicalURL: resolveURLPtr(firstLinkHrefByRelToken(doc, "canonical"), baseURL),
		Title:        firstTextBySelectors(doc.Selection, []string{"h1"}, 8, 240),
		Byline: cleanByline(
			valueOrEmpty(firstTextBySelectors(doc.Selection, []string{
				"[rel='author']",
				"[itemprop='author']",
				".byline",
				".article-byline",
				".post-author",
				".author-name",
				"[class*='author']",
				"[class*='byline']",
				"[id*='author']",
				"[id*='byline']",
			}, 3, 120)),
		),
		Excerpt: firstParagraphExcerpt(doc.Selection),
		LeadImage: firstImageURL(
			doc.Find("article, main, [role='main'], body").First(),
			baseURL,
		),
		Tags:       findTagLinks(doc),
		Categories: findCategoryLinks(doc),
	}
}

func extractDOMFallbackTimes(doc *goquery.Document, preferredRoot *goquery.Selection) types.Article {
	return types.Article{
		PublishedAt: findPublishedTime(doc, preferredRoot),
		UpdatedAt:   findUpdatedTime(doc, preferredRoot),
	}
}

func firstLinkHrefByRelToken(doc *goquery.Document, token string) string {
	if doc == nil {
		return ""
	}

	found := ""
	doc.Find("link[rel]").EachWithBreak(func(_ int, sel *goquery.Selection) bool {
		if !hasSpaceSeparatedToken(firstAttr(sel, "rel"), token) {
			return true
		}

		found = firstAttr(sel, "href")

		return found == ""
	})

	return found
}

func fillTitleAndSiteFromText(titleText string, article *types.Article) {
	titleText = normalizeWhitespace(titleText)
	if titleText == "" {
		return
	}

	derivedTitle, derivedSite := splitDocumentTitle(titleText)
	if article.Title == nil {
		article.Title = stringPtr(derivedTitle)
	}

	if article.SiteName == nil {
		if article.Title != nil {
			if _, matchedSite := splitDocumentTitleAgainstKnownTitle(titleText, *article.Title); matchedSite != "" {
				derivedSite = matchedSite
			}
		}

		article.SiteName = stringPtr(derivedSite)
	}
}

func splitDocumentTitle(title string) (string, string) {
	title = normalizeWhitespace(title)
	if title == "" {
		return "", ""
	}

	separators := []string{" | ", " - ", " — ", " :: ", " • "}
	for _, separator := range separators {
		parts := strings.Split(title, separator)
		if len(parts) < 2 {
			continue
		}

		left := normalizeWhitespace(parts[0])
		right := normalizeWhitespace(parts[len(parts)-1])
		if left == "" || right == "" {
			continue
		}

		if len(left) >= len(right) {
			return left, right
		}

		return right, left
	}

	return title, ""
}

func splitDocumentTitleAgainstKnownTitle(raw string, known string) (string, string) {
	separators := []string{" | ", " - ", " — ", " :: ", " • "}
	for _, separator := range separators {
		parts := strings.Split(raw, separator)
		if len(parts) < 2 {
			continue
		}

		for _, part := range parts {
			part = normalizeWhitespace(part)
			if !titleMatches(part, known) {
				continue
			}

			for _, other := range parts {
				other = normalizeWhitespace(other)
				if other != "" && !titleMatches(other, known) {
					return part, other
				}
			}
		}
	}

	return known, ""
}

func findPublishedTime(doc *goquery.Document, preferredRoot *goquery.Selection) *string {
	return findScopedTime(doc, preferredRoot, false)
}

func findUpdatedTime(doc *goquery.Document, preferredRoot *goquery.Selection) *string {
	return findScopedTime(doc, preferredRoot, true)
}

func findScopedTime(doc *goquery.Document, preferredRoot *goquery.Selection, requireUpdated bool) *string {
	if doc == nil {
		return nil
	}

	if found := findTimeInRoot(preferredRoot, requireUpdated); found != nil {
		return found
	}

	for _, selector := range []string{"article", "main", "[role='main']"} {
		var found *string
		doc.Find(selector).EachWithBreak(func(_ int, root *goquery.Selection) bool {
			found = findTimeInRoot(root, requireUpdated)

			return found == nil
		})

		if found != nil {
			return found
		}
	}

	return findTimeInRoot(doc.Selection, requireUpdated)
}

func findTimeInRoot(root *goquery.Selection, requireUpdated bool) *string {
	if root == nil || root.Length() == 0 {
		return nil
	}

	var found *string
	root.Find("time").EachWithBreak(func(_ int, sel *goquery.Selection) bool {
		isUpdated := hasKeyword(classID(sel), []string{"update", "modified"})
		if requireUpdated != isUpdated {
			return true
		}

		if timestamp := normalizeTimestampStrict(firstAttr(sel, "datetime")); timestamp != nil {
			found = timestamp

			return false
		}

		if timestamp := normalizeTimestampStrict(sel.Text()); timestamp != nil {
			found = timestamp

			return false
		}

		return true
	})

	return found
}

func firstParagraphExcerpt(root *goquery.Selection) *string {
	var excerpt *string

	root.Find("article p, main p, [role='main'] p, p").EachWithBreak(func(_ int, sel *goquery.Selection) bool {
		text := normalizeWhitespace(sel.Text())
		if len(text) < 48 {
			return true
		}

		excerpt = stringPtr(truncateText(text, 220))

		return false
	})

	return excerpt
}

func findTagLinks(doc *goquery.Document) []string {
	values := make([]string, 0)

	doc.Find("a[rel='tag'], .tags a, [class*='tag'] a, a[href*='/tag/']").Each(func(_ int, sel *goquery.Selection) {
		text := normalizeWhitespace(sel.Text())
		if text != "" && len(text) <= 64 {
			values = append(values, text)
		}
	})

	return uniqueStrings(values)
}

func findCategoryLinks(doc *goquery.Document) []string {
	values := make([]string, 0)

	doc.Find(".category a, [class*='category'] a, a[href*='/category/']").Each(func(_ int, sel *goquery.Selection) {
		text := normalizeWhitespace(sel.Text())
		if text != "" && len(text) <= 64 {
			values = append(values, text)
		}
	})

	return uniqueStrings(values)
}
