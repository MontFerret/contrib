package logutil

import "github.com/rs/zerolog"

func WithComponent(ctx zerolog.Context, name string) zerolog.Context {
	return ctx.Str("component", name)
}
