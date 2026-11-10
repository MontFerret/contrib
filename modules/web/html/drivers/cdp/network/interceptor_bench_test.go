package network_test

import (
	"testing"

	cdpnetwork "github.com/mafredri/cdp/protocol/network"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
)

func BenchmarkInterceptorFilter(b *testing.B) {
	filter, err := network.NewInterceptorFilter(&network.Filter{
		Patterns: []drivers.ResourceFilter{{Type: "Image", URL: "*.png"}},
	})
	if err != nil {
		b.Fatal(err)
	}

	request := cdpnetwork.Request{URL: "https://example.com/image.png"}
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if !filter.Filter(cdpnetwork.ResourceTypeImage, request) {
			b.Fatal("image was not filtered")
		}
	}
}
