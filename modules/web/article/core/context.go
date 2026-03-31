package core

import "context"

type extractorContextKey struct{}

func WithExtractor(ctx context.Context, extractor *Extractor) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, extractorContextKey{}, extractor)
}

func ExtractorFromContext(ctx context.Context) *Extractor {
	if ctx == nil {
		return nil
	}

	value, ok := ctx.Value(extractorContextKey{}).(*Extractor)
	if !ok {
		return nil
	}

	return value
}
