package core

import (
	"context"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestParseDecodeOptions(t *testing.T) {
	ctx := context.Background()

	t.Run("valid options", func(t *testing.T) {
		input := runtime.NewObjectWith(map[string]runtime.Value{
			"datetime": runtime.NewString(DecodeDateTimeNative),
			"strict":   runtime.False,
		})

		opts, err := ParseDecodeOptions(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opts.DateTime != DecodeDateTimeNative || opts.Strict {
			t.Fatalf("unexpected options: %+v", opts)
		}
	})

	t.Run("unknown option", func(t *testing.T) {
		input := runtime.NewObjectWith(map[string]runtime.Value{
			"extra": runtime.True,
		})

		_, err := ParseDecodeOptions(ctx, input)
		if err == nil {
			t.Fatal("expected unknown option error")
		}

		if !strings.Contains(err.Error(), `unknown decode option "extra"`) {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestParseEncodeOptions(t *testing.T) {
	ctx := context.Background()

	t.Run("valid options", func(t *testing.T) {
		input := runtime.NewObjectWith(map[string]runtime.Value{
			"sort_keys": runtime.True,
			"datetime":  runtime.NewString(EncodeDateTimePreserve),
		})

		opts, err := ParseEncodeOptions(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !opts.SortKeys || opts.DateTime != EncodeDateTimePreserve {
			t.Fatalf("unexpected options: %+v", opts)
		}
	})

	t.Run("invalid datetime mode", func(t *testing.T) {
		input := runtime.NewObjectWith(map[string]runtime.Value{
			"datetime": runtime.NewString("bad"),
		})

		_, err := ParseEncodeOptions(ctx, input)
		if err == nil {
			t.Fatal("expected invalid encode datetime error")
		}
	})
}
