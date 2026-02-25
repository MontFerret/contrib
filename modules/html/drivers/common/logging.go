package common

import "github.com/rs/zerolog"

func LoggerWithName(ctx zerolog.Context, name string) zerolog.Context {
	return ctx.Str("component", name)
}
