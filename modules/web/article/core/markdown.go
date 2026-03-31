package core

import (
	"context"
	"net/url"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
)

type markdownConverterKey struct{}

func newMarkdownConverter() *converter.Converter {
	return converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			table.NewTablePlugin(
				table.WithHeaderPromotion(true),
			),
		),
	)
}

func WithExtractor(ctx context.Context, extractor *Extractor) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, markdownConverterKey{}, extractor)
}

func ExtractorFromContext(ctx context.Context) *Extractor {
	if ctx == nil {
		return nil
	}

	value, ok := ctx.Value(markdownConverterKey{}).(*Extractor)
	if !ok {
		return nil
	}

	return value
}

func (e *Extractor) markdownConverterOrNew() *converter.Converter {
	if e != nil && e.markdownConverter != nil {
		return e.markdownConverter
	}

	return newMarkdownConverter()
}

func markdownConvertOptions(baseURL *url.URL) []converter.ConvertOptionFunc {
	if baseURL == nil {
		return nil
	}

	return []converter.ConvertOptionFunc{
		markdownDomainOption(baseURL),
	}
}

func markdownDomainOption(baseURL *url.URL) converter.ConvertOptionFunc {
	return converter.WithDomain(baseURL.String())
}
