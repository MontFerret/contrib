package core

import "testing"

func BenchmarkExtractLargeCandidatePage(b *testing.B) {
	fixture := buildLargeCandidateFixture(180)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		article := Extract(fixture)
		if article.Text == nil {
			b.Fatal("expected article text")
		}
	}
}
