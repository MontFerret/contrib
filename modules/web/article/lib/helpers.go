package lib

import (
	"context"
	"html"
	"net/url"
	"sort"
	"strings"

	"github.com/MontFerret/contrib/modules/web/article/core"
	"github.com/MontFerret/contrib/modules/web/article/types"
	htmldrivers "github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func extractArticle(ctx context.Context, args ...runtime.Value) (types.Article, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return types.Article{}, err
	}

	source, err := normalizeSource(ctx, args[0])
	if err != nil {
		return types.Article{}, err
	}

	return core.ExtractSource(ctx, source), nil
}

func normalizeSource(ctx context.Context, value runtime.Value) (core.Source, error) {
	switch input := value.(type) {
	case runtime.String:
		return core.Source{
			HTML: input.String(),
		}, nil
	default:
		doc, err := htmldrivers.ToDocument(value)
		if err == nil {
			return sourceFromDocument(ctx, doc)
		}

		el, err := htmldrivers.ToElement(value)
		if err == nil {
			return sourceFromElement(ctx, el)
		}

		return core.Source{}, runtime.TypeErrorOf(
			value,
			runtime.TypeString,
			htmldrivers.HTMLPageType,
			htmldrivers.HTMLDocumentType,
			htmldrivers.HTMLElementType,
		)
	}
}

func sourceFromDocument(ctx context.Context, doc htmldrivers.HTMLDocument) (core.Source, error) {
	htmlValue, err := snapshotElementHTML(ctx, doc.GetElement())
	if err != nil {
		return core.Source{}, err
	}

	source := core.Source{
		HTML:      htmlValue,
		SourceURL: parseSourceURL(doc.GetURL().String()),
		TitleHint: optionalRuntimeString(doc.GetTitle()),
	}

	return source, nil
}

func sourceFromElement(ctx context.Context, el htmldrivers.HTMLElement) (core.Source, error) {
	htmlValue, err := snapshotElementHTML(ctx, el)
	if err != nil {
		return core.Source{}, err
	}

	return core.Source{HTML: htmlValue}, nil
}

func snapshotElementHTML(ctx context.Context, el htmldrivers.HTMLElement) (string, error) {
	innerHTML, err := el.GetInnerHTML(ctx)
	if err != nil {
		return "", err
	}

	nodeName, err := el.GetNodeName(ctx)
	if err != nil {
		return "", err
	}

	tagName := strings.TrimSpace(nodeName.String())
	if tagName == "" || strings.HasPrefix(tagName, "#") {
		return innerHTML.String(), nil
	}

	attrs, err := el.GetAttributes(ctx)
	if err != nil {
		return "", err
	}

	attrString, err := serializeAttributes(ctx, attrs)
	if err != nil {
		return "", err
	}

	return "<" + tagName + attrString + ">" + innerHTML.String() + "</" + tagName + ">", nil
}

func serializeAttributes(ctx context.Context, attrs runtime.Map) (string, error) {
	if attrs == nil {
		return "", nil
	}

	values := make(map[string]string)
	err := attrs.ForEach(ctx, func(_ context.Context, key, value runtime.Value) (runtime.Boolean, error) {
		values[key.String()] = value.String()

		return runtime.True, nil
	})
	if err != nil {
		return "", err
	}

	if len(values) == 0 {
		return "", nil
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	var builder strings.Builder
	for _, key := range keys {
		builder.WriteByte(' ')
		builder.WriteString(key)
		builder.WriteString(`="`)
		builder.WriteString(html.EscapeString(values[key]))
		builder.WriteByte('"')
	}

	return builder.String(), nil
}

func parseSourceURL(raw string) *url.URL {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return nil
	}

	return parsed
}

func optionalRuntimeString(value runtime.String) *string {
	text := strings.TrimSpace(value.String())
	if text == "" {
		return nil
	}

	return &text
}
