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

func NewMarkdownConverter() *converter.Converter {
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

func WithMarkdownConverter(ctx context.Context, conv *converter.Converter) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, markdownConverterKey{}, conv)
}

func MarkdownConverterFromContext(ctx context.Context) *converter.Converter {
	if ctx == nil {
		return nil
	}

	value, ok := ctx.Value(markdownConverterKey{}).(*converter.Converter)
	if !ok {
		return nil
	}

	return value
}

func resolveMarkdownConverter(ctx context.Context) *converter.Converter {
	conv := MarkdownConverterFromContext(ctx)
	if conv != nil {
		return conv
	}

	return NewMarkdownConverter()
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
