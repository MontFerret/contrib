package core

import (
	"net/url"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
)

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
