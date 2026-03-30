package core

import (
	"fmt"
	"testing"
)

func BenchmarkMatchManyRules(b *testing.B) {
	allow := make([]string, 0, 500)
	disallow := make([]string, 0, 500)

	for i := 0; i < 500; i++ {
		allow = append(allow, fmt.Sprintf("/public/%03d/*", i))
		disallow = append(disallow, fmt.Sprintf("/section/%03d/private", i))
	}

	doc := Document{
		Groups: []Group{
			{
				UserAgents: []string{"*"},
				Allow:      allow,
				Disallow:   disallow,
			},
		},
	}

	path := "/section/499/private/page"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := Match(doc, path, "BenchmarkBot")
		if result.Allowed {
			b.Fatalf("expected path to be disallowed, got %+v", result)
		}
	}
}
