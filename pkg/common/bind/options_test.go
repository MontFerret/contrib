package bind

import (
	"context"
	"errors"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type testOptions struct {
	Delimiter  string   `json:"delimiter"`
	NullValues []string `json:"nullValues"`
	Header     bool     `json:"header"`
}

func TestDecodeMapArgOrDefault(t *testing.T) {
	ctx := context.Background()
	defaults := testOptions{
		Delimiter: ",",
		Header:    true,
	}

	t.Run("missing arg returns defaults", func(t *testing.T) {
		got, err := DecodeMapArgOrDefault(ctx, nil, 0, defaults)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Delimiter != defaults.Delimiter || got.Header != defaults.Header || len(got.NullValues) != 0 {
			t.Fatalf("unexpected defaults: %#v", got)
		}
	})

	t.Run("decodes provided map", func(t *testing.T) {
		args := []runtime.Value{
			runtime.NewString("data"),
			runtime.NewObjectWith(map[string]runtime.Value{
				"delimiter":  runtime.NewString(";"),
				"header":     runtime.False,
				"nullValues": runtime.NewArrayWith(runtime.NewString("null")),
			}),
		}

		got, err := DecodeMapArgOrDefault(ctx, args, 1, defaults)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Delimiter != ";" || got.Header {
			t.Fatalf("unexpected decoded options: %#v", got)
		}
		if len(got.NullValues) != 1 || got.NullValues[0] != "null" {
			t.Fatalf("unexpected decoded null values: %#v", got.NullValues)
		}
	})

	t.Run("wrong arg type", func(t *testing.T) {
		_, err := DecodeMapArgOrDefault(ctx, []runtime.Value{runtime.NewString("bad")}, 0, defaults)
		if !errors.Is(err, runtime.ErrInvalidType) {
			t.Fatalf("expected invalid type error, got %v", err)
		}
	})
}
