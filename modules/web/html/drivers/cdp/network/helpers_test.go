package network

import (
	"testing"
	"time"

	cdpnetwork "github.com/mafredri/cdp/protocol/network"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

func TestToDriverRequestPrefersPostDataEntries(t *testing.T) {
	legacy := "legacy"
	part1 := "hello "
	part2 := "world"

	req := cdpnetwork.Request{
		URL:             "https://example.com",
		Method:          "POST",
		PostData:        &legacy,
		PostDataEntries: []cdpnetwork.PostDataEntry{{Bytes: &part1}, {Bytes: &part2}},
	}

	driverReq := toDriverRequest(req)

	if string(driverReq.Body) != "hello world" {
		t.Fatalf("expected request body from postDataEntries, got %q", string(driverReq.Body))
	}
}

func TestToDriverRequestFallsBackToLegacyPostData(t *testing.T) {
	legacy := "legacy"

	req := cdpnetwork.Request{
		URL:      "https://example.com",
		Method:   "POST",
		PostData: &legacy,
	}

	driverReq := toDriverRequest(req)

	if string(driverReq.Body) != legacy {
		t.Fatalf("expected request body from legacy postData, got %q", string(driverReq.Body))
	}
}

func TestFromDriverCookieOmitsUnsetOptionalFields(t *testing.T) {
	param := fromDriverCookie("example.com/path", drivers.HTTPCookie{
		Name:  "session",
		Value: "abc123",
	})

	if param.URL == nil || *param.URL != "http://example.com/path" {
		t.Fatalf("expected normalized URL, got %#v", param.URL)
	}

	if param.Path != nil {
		t.Fatalf("expected empty path to be omitted, got %#v", *param.Path)
	}

	if param.Domain != nil {
		t.Fatalf("expected empty domain to be omitted, got %#v", *param.Domain)
	}

	if param.Secure != nil {
		t.Fatalf("expected secure=false to be omitted")
	}

	if param.HTTPOnly != nil {
		t.Fatalf("expected httpOnly=false to be omitted")
	}

	if param.Expires != 0 {
		t.Fatalf("expected session cookie expiry to be omitted, got %v", param.Expires)
	}
}

func TestFromDriverCookieUsesMaxAgeAsExpiry(t *testing.T) {
	before := time.Now()
	param := fromDriverCookie("https://example.com", drivers.HTTPCookie{
		Name:   "session",
		Value:  "abc123",
		MaxAge: 10,
	})
	after := time.Now()

	if param.Expires == 0 {
		t.Fatalf("expected expiry to be set from maxAge")
	}

	min := before.Add(10 * time.Second).Unix()
	max := after.Add(10 * time.Second).Unix()
	actual := param.Expires.Time().Unix()

	if actual < min || actual > max {
		t.Fatalf("expected expiry in [%d, %d], got %d", min, max, actual)
	}
}

func TestFromDriverCookieDeleteOmitsEmptyDomainAndPath(t *testing.T) {
	args := fromDriverCookieDelete("example.com", drivers.HTTPCookie{
		Name: "session",
	})

	if args.URL == nil || *args.URL != "http://example.com" {
		t.Fatalf("expected normalized URL, got %#v", args.URL)
	}

	if args.Path != nil {
		t.Fatalf("expected empty path to be omitted, got %#v", *args.Path)
	}

	if args.Domain != nil {
		t.Fatalf("expected empty domain to be omitted, got %#v", *args.Domain)
	}
}
