package dom

import (
	"context"
	"testing"
)

func BenchmarkDecodeDOMEventPayload(b *testing.B) {
	const payload = `{"type":"click","target":{"id":"submit","className":"button"},"clientX":120,"clientY":80}`
	ctx := context.Background()
	b.ReportAllocs()

	for b.Loop() {
		if _, err := decodeDOMEventPayload(ctx, payload); err != nil {
			b.Fatal(err)
		}
	}
}
