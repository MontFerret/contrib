package core

import "github.com/MontFerret/ferret/v2/pkg/runtime"

const (
	TypeURLSet       = "urlset"
	TypeSitemapIndex = "sitemapindex"
)

// URLEntry represents a single sitemap URL entry.
type URLEntry struct {
	Loc        string
	LastMod    string
	ChangeFreq string
	Priority   *float64
	Source     string
}

// SitemapRef represents a sitemap reference from a sitemap index.
type SitemapRef struct {
	Loc     string
	LastMod string
}

// Document represents a parsed sitemap document.
type Document struct {
	Type     string
	URLs     []URLEntry
	Sitemaps []SitemapRef
}

func (e URLEntry) ToValue() runtime.Value {
	props := map[string]runtime.Value{
		"loc":        runtime.NewString(e.Loc),
		"lastmod":    optionalStringValue(e.LastMod),
		"changefreq": optionalStringValue(e.ChangeFreq),
		"priority":   runtime.None,
		"source":     runtime.NewString(e.Source),
	}

	if e.Priority != nil {
		props["priority"] = runtime.NewFloat(*e.Priority)
	}

	return runtime.NewObjectWith(props)
}

func (s SitemapRef) ToValue() runtime.Value {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"loc":     runtime.NewString(s.Loc),
		"lastmod": optionalStringValue(s.LastMod),
	})
}

func (d Document) ToValue() runtime.Value {
	switch d.Type {
	case TypeURLSet:
		return runtime.NewObjectWith(map[string]runtime.Value{
			"type": runtime.NewString(TypeURLSet),
			"urls": urlEntriesToArray(d.URLs),
		})
	case TypeSitemapIndex:
		return runtime.NewObjectWith(map[string]runtime.Value{
			"type":     runtime.NewString(TypeSitemapIndex),
			"sitemaps": sitemapRefsToArray(d.Sitemaps),
		})
	default:
		return runtime.NewObjectWith(map[string]runtime.Value{
			"type": runtime.NewString(d.Type),
		})
	}
}

func URLEntriesToValue(entries []URLEntry) runtime.Value {
	return urlEntriesToArray(entries)
}

func urlEntriesToArray(entries []URLEntry) *runtime.Array {
	values := make([]runtime.Value, 0, len(entries))

	for _, entry := range entries {
		values = append(values, entry.ToValue())
	}

	return runtime.NewArrayOf(values)
}

func sitemapRefsToArray(entries []SitemapRef) *runtime.Array {
	values := make([]runtime.Value, 0, len(entries))

	for _, entry := range entries {
		values = append(values, entry.ToValue())
	}

	return runtime.NewArrayOf(values)
}

func optionalStringValue(input string) runtime.Value {
	if input == "" {
		return runtime.None
	}

	return runtime.NewString(input)
}
