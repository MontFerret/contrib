package common

import (
	"bytes"
	"context"
	"strconv"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/gorilla/css/scanner"
)

func DeserializeStyles(ctx context.Context, input runtime.String) (runtime.Map, error) {
	styles := runtime.NewObject()

	if input == runtime.EmptyString {
		return styles, nil
	}

	s := scanner.New(input.String())

	var name string
	var value bytes.Buffer
	var setValue = func() {
		_ = styles.Set(ctx, runtime.NewString(strings.TrimSpace(name)), runtime.NewString(strings.TrimSpace(value.String())))
		name = ""
		value.Reset()
	}

	for {
		token := s.Next()

		if token == nil {
			break
		}

		if token.Type == scanner.TokenEOF {
			break
		}

		if name == "" && token.Type == scanner.TokenIdent {
			name = token.Value

			// skip : and white spaces
			for {
				token = s.Next()

				if token.Value != ":" && token.Type != scanner.TokenS {
					break
				}
			}
		}

		switch token.Type {
		case scanner.TokenChar:
			// end of style declaration
			if token.Value == ";" {
				if name != "" {
					setValue()
				}
			} else {
				value.WriteString(token.Value)
			}
		case scanner.TokenNumber:
			num, err := strconv.ParseFloat(token.Value, 64)

			if err == nil {
				_ = styles.Set(ctx, runtime.NewString(name), runtime.NewFloat(num))
				// reset prop
				name = ""
				value.Reset()
			}
		default:
			value.WriteString(token.Value)
		}
	}

	if name != "" && value.Len() > 0 {
		setValue()
	}

	return styles, nil
}

func SerializeStyles(ctx context.Context, styles runtime.Map) (runtime.String, error) {
	if styles == nil {
		return runtime.EmptyString, nil
	}

	var b bytes.Buffer

	err := styles.ForEach(ctx, func(ctx context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		b.WriteString(key.String())
		b.WriteString(": ")
		b.WriteString(value.String())
		b.WriteString("; ")

		return true, nil
	})

	if err != nil {
		return runtime.EmptyString, err
	}

	return runtime.NewString(b.String()), nil
}
