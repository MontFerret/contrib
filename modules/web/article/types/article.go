package types

import "github.com/MontFerret/ferret/v2/pkg/runtime"

// Article is the normalized WEB::ARTICLE extraction result.
type Article struct {
	Title              *string
	Byline             *string
	Excerpt            *string
	SiteName           *string
	PublishedAt        *string
	UpdatedAt          *string
	Lang               *string
	Dir                *string
	CanonicalURL       *string
	LeadImage          *string
	Text               *string
	HTML               *string
	Markdown           *string
	WordCount          *int
	ReadingTimeMinutes *int
	Tags               []string
	Categories         []string
}

// ToValue encodes the article into a Ferret object with explicit nulls.
func (a Article) ToValue() runtime.Value {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"title":              optionalStringValue(a.Title),
		"byline":             optionalStringValue(a.Byline),
		"excerpt":            optionalStringValue(a.Excerpt),
		"siteName":           optionalStringValue(a.SiteName),
		"publishedAt":        optionalStringValue(a.PublishedAt),
		"updatedAt":          optionalStringValue(a.UpdatedAt),
		"lang":               optionalStringValue(a.Lang),
		"dir":                optionalStringValue(a.Dir),
		"canonicalUrl":       optionalStringValue(a.CanonicalURL),
		"leadImage":          optionalStringValue(a.LeadImage),
		"text":               optionalStringValue(a.Text),
		"html":               optionalStringValue(a.HTML),
		"markdown":           optionalStringValue(a.Markdown),
		"wordCount":          optionalIntValue(a.WordCount),
		"readingTimeMinutes": optionalIntValue(a.ReadingTimeMinutes),
		"tags":               optionalStringsValue(a.Tags),
		"categories":         optionalStringsValue(a.Categories),
	})
}

func optionalStringValue(value *string) runtime.Value {
	if value == nil {
		return runtime.None
	}

	return runtime.NewString(*value)
}

func optionalIntValue(value *int) runtime.Value {
	if value == nil {
		return runtime.None
	}

	return runtime.NewInt64(int64(*value))
}

func optionalStringsValue(values []string) runtime.Value {
	if len(values) == 0 {
		return runtime.None
	}

	items := make([]runtime.Value, 0, len(values))
	for _, value := range values {
		items = append(items, runtime.NewString(value))
	}

	return runtime.NewArrayOf(items)
}
