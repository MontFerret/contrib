package core

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func BenchmarkDecodeSchema(b *testing.B) {
	value := runtime.NewObjectWith(map[string]runtime.Value{
		"type": runtime.NewString("object"),
		"properties": runtime.NewObjectWith(map[string]runtime.Value{
			"name": runtime.NewObjectWith(map[string]runtime.Value{
				"type": runtime.NewString("string"),
			}),
		}),
	})
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if _, err := DecodeSchema(ctx, value); err != nil {
			b.Fatal(err)
		}
	}
}
