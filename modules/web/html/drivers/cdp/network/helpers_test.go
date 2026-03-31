package network

import (
	"testing"

	cdpnetwork "github.com/mafredri/cdp/protocol/network"
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
