package cdp

import (
	"context"
	"strings"
	"testing"

	cdpnetwork "github.com/mafredri/cdp/protocol/network"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
)

func TestAllowedOriginsOption(t *testing.T) {
	values := []string{"https://example.com"}
	option := WithAllowedOrigins(values...)
	values[0] = "https://evil.test"
	driver := New(option)
	filter, err := network.NewInterceptorFilter(driver.policy.filter)
	if err != nil {
		t.Fatal(err)
	}

	if filter.Filter(cdpnetwork.ResourceTypeDocument, cdpnetwork.Request{URL: "https://example.com/"}) {
		t.Fatal("option did not copy the original list")
	}

	if !filter.Filter(cdpnetwork.ResourceTypeDocument, cdpnetwork.Request{URL: values[0]}) {
		t.Fatal("caller changed the policy through its original slice")
	}

	empty := New(WithAllowedOrigins("https://example.com"), WithAllowedOrigins())
	if empty.policy == nil || empty.policy.filter.AllowedOrigins == nil || len(empty.policy.filter.AllowedOrigins) != 0 {
		t.Fatal("explicit empty policy was lost or merged")
	}

	if New().policy != nil {
		t.Fatal("unrestricted driver created policy state")
	}
}

func TestInvalidPolicyFailsBeforeConnecting(t *testing.T) {
	driver := New(WithAddress("http://127.0.0.1:1"), WithAllowedOrigins("https://user:secret@example.com/path"))
	if _, err := driver.Open(context.Background(), drivers.Params{}); err == nil || !strings.Contains(err.Error(), "allowed origin") || strings.Contains(err.Error(), "secret") {
		t.Fatalf("Open error = %v", err)
	}

	if _, err := driver.Parse(context.Background(), drivers.ParseParams{}); err == nil || !strings.Contains(err.Error(), "allowed origin") {
		t.Fatalf("Parse error = %v", err)
	}

	if err := driver.Close(); err != nil {
		t.Fatal(err)
	}

	if err := driver.Close(); err != nil {
		t.Fatal(err)
	}
}
