package network_test

import (
	"fmt"
	"strings"
	"testing"

	cdpnetwork "github.com/mafredri/cdp/protocol/network"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
)

func TestAllowedOrigins(t *testing.T) {
	tests := []struct {
		name    string
		origins []string
		url     string
		blocked bool
	}{
		{"omitted", nil, "https://other.example/", false},
		{"empty", []string{}, "https://example.com/", true},
		{"blank initialization", []string{}, "about:blank", false},
		{"blank query is not exception", []string{}, "about:blank?url=https://example.com", true},
		{"host casing and default port", []string{"https://EXAMPLE.com:443/"}, "https://example.com/a?b=c#d", false},
		{"HTTP default port", []string{"http://example.com"}, "http://EXAMPLE.com:80/", false},
		{"scheme casing", []string{"HTTPS://EXAMPLE.COM"}, "https://example.com/", false},
		{"other scheme", []string{"https://example.com"}, "http://example.com/", true},
		{"other port", []string{"https://example.com"}, "https://example.com:444/", true},
		{"subdomain", []string{"https://example.com"}, "https://sub.example.com/", true},
		{"suffix", []string{"https://example.com"}, "https://example.com.evil.test/", true},
		{"userinfo cannot disguise host", []string{"https://example.com"}, "https://example.com@evil.test/", true},
		{"request userinfo preserves origin", []string{"https://example.com"}, "https://user:pass@example.com/", false},
		{"IPv6", []string{"http://[2001:db8:0:0::1]:80"}, "http://[2001:db8::1]/", false},
		{"IPv4", []string{"http://127.0.0.1:8080"}, "http://127.0.0.1:8080/x", false},
		{"IDNA", []string{"https://BÜCHER.example"}, "https://xn--bcher-kva.example/", false},
		{"trailing dot remains distinct", []string{"https://example.com"}, "https://example.com./", true},
		{"duplicate origins", []string{"https://example.com", "https://example.com:443"}, "https://example.com/", false},
		{"file", []string{}, "file:///etc/passwd", true},
		{"data", []string{}, "data:text/plain,hello", true},
		{"malformed request", []string{}, "://broken", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter, err := network.NewInterceptorFilter(&network.Filter{AllowedOrigins: test.origins})
			if err != nil {
				t.Fatal(err)
			}

			if actual := filter.Filter(cdpnetwork.ResourceTypeDocument, cdpnetwork.Request{URL: test.url}); actual != test.blocked {
				t.Fatalf("blocked = %v, want %v", actual, test.blocked)
			}
		})
	}
}

func TestInvalidAllowedOrigins(t *testing.T) {
	values := []string{
		"", "example.com", "//example.com", "/relative", "https:", "https:///path", "https://",
		"ws://example.com", "file:///tmp/example", "about:blank", "https://user:secret@example.com",
		"https://*.example.com", "https://example.com/path", "https://example.com/%2f",
		"https://example.com?", "https://example.com?query", "https://example.com#", "https://example.com#fragment",
		"https://example.com:", "https://example.com:-1", "https://example.com:65536", "https://example.com:port",
		"https://[::1", "https://[127.0.0.1]", "https://[example.com]", "https://[fe80::1%25en0]", "https://bad host", " https://example.com",
		"https://example.com\\evil", "https://bad_host.example", "https://a..example", "https://.",
	}

	for index, value := range values {
		t.Run(fmt.Sprintf("case_%d", index), func(t *testing.T) {
			_, err := network.NewInterceptorFilter(&network.Filter{AllowedOrigins: []string{value}})
			if err == nil {
				t.Fatalf("accepted invalid origin %q", value)
			}

			if strings.Contains(err.Error(), "secret") {
				t.Fatal("configuration error exposed credentials")
			}
		})
	}
}

func TestOriginFilterCopiesConfiguration(t *testing.T) {
	origins := []string{"https://example.com"}
	filter, err := network.NewInterceptorFilter(&network.Filter{AllowedOrigins: origins})
	if err != nil {
		t.Fatal(err)
	}

	origins[0] = "https://evil.test"
	if !filter.Filter(cdpnetwork.ResourceTypeFetch, cdpnetwork.Request{URL: origins[0]}) {
		t.Fatal("configuration mutation changed compiled permissions")
	}
}

func BenchmarkAllowedOrigins(b *testing.B) {
	for _, size := range []int{1, 100} {
		origins := make([]string, size)
		for index := range origins {
			origins[index] = fmt.Sprintf("https://host%d.example", index)
		}

		filter, err := network.NewInterceptorFilter(&network.Filter{AllowedOrigins: origins})
		if err != nil {
			b.Fatal(err)
		}

		for _, path := range []string{origins[0] + "/resource", "https://denied.example/resource"} {
			b.Run(fmt.Sprintf("%d/%s", size, path), func(b *testing.B) {
				request := cdpnetwork.Request{URL: path}
				b.ReportAllocs()

				for b.Loop() {
					filter.Filter(cdpnetwork.ResourceTypeFetch, request)
				}
			})
		}
	}
}
