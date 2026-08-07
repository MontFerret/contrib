package drivers_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/mafredri/cdp/protocol/page"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp"
	cdpdom "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/dom"
	cdpnetwork "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
	"github.com/MontFerret/contrib/modules/web/html/drivers/memory"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

var (
	_ runtime.Comparable = (*drivers.HTTPHeaders)(nil)
	_ runtime.Equatable  = (*drivers.HTTPHeaders)(nil)
	_ runtime.Comparable = (*drivers.HTTPRequest)(nil)
	_ runtime.Equatable  = (*drivers.HTTPRequest)(nil)
	_ runtime.Comparable = (*drivers.HTTPResponse)(nil)
	_ runtime.Equatable  = (*drivers.HTTPResponse)(nil)
	_ runtime.Comparable = drivers.HTTPCookie{}
	_ runtime.Equatable  = drivers.HTTPCookie{}
	_ runtime.Comparable = (*memory.HTMLPage)(nil)
	_ runtime.Equatable  = (*memory.HTMLPage)(nil)
	_ runtime.Comparable = (*memory.HTMLDocument)(nil)
	_ runtime.Equatable  = (*memory.HTMLDocument)(nil)
	_ runtime.Comparable = (*memory.HTMLElement)(nil)
	_ runtime.Equatable  = (*memory.HTMLElement)(nil)
	_ runtime.Comparable = (*cdp.HTMLPage)(nil)
	_ runtime.Equatable  = (*cdp.HTMLPage)(nil)
	_ runtime.Comparable = (*cdpdom.HTMLDocument)(nil)
	_ runtime.Equatable  = (*cdpdom.HTMLDocument)(nil)
	_ runtime.Comparable = (*cdpdom.HTMLElement)(nil)
	_ runtime.Equatable  = (*cdpdom.HTMLElement)(nil)
	_ runtime.Comparable = cdpdom.FrameID("")
	_ runtime.Equatable  = cdpdom.FrameID("")
	_ runtime.Comparable = (*cdpnetwork.NavigationEvent)(nil)
	_ runtime.Equatable  = (*cdpnetwork.NavigationEvent)(nil)
)

func TestDriverValueComparisonContracts(t *testing.T) {
	t.Parallel()

	headers := drivers.NewHTTPHeadersWith(map[string][]string{
		"Accept": {"application/json"},
		"X-Test": {"one", "two"},
	})
	equalHeaders := drivers.NewHTTPHeadersWith(map[string][]string{
		"X-Test": {"one", "two"},
		"Accept": {"application/json"},
	})
	assertEqualContract(t, headers, equalHeaders)

	cookie := drivers.HTTPCookie{Name: "session", Value: "abc", Path: "/"}
	assertEqualContract(t, cookie, cookie.Copy())

	request := &drivers.HTTPRequest{
		URL:     "https://example.com",
		Method:  "POST",
		Headers: headers,
		Body:    []byte("first"),
	}
	equalRequest := &drivers.HTTPRequest{
		URL:     request.URL,
		Method:  request.Method,
		Headers: equalHeaders,
		Body:    []byte("second"),
	}
	assertEqualContract(t, request, equalRequest)

	response := &drivers.HTTPResponse{Headers: headers, StatusCode: 200, URL: "https://one.example"}
	equalResponse := &drivers.HTTPResponse{Headers: equalHeaders, StatusCode: 200, URL: "https://two.example"}
	assertEqualContract(t, response, equalResponse)

	frameID := cdpdom.FrameID("frame")
	assertEqualContract(t, frameID, frameID.Copy())

	leftEvent := &cdpnetwork.NavigationEvent{URL: "https://example.com", FrameID: page.FrameID(frameID), MimeType: "text/html"}
	rightEvent := &cdpnetwork.NavigationEvent{URL: leftEvent.URL, FrameID: leftEvent.FrameID, MimeType: "application/xhtml+xml"}
	assertEqualContract(t, leftEvent, rightEvent)
}

func TestHTTPResponseComparisonUsesOtherStatusCode(t *testing.T) {
	t.Parallel()

	left := &drivers.HTTPResponse{Headers: drivers.NewHTTPHeaders(), StatusCode: 200}
	right := &drivers.HTTPResponse{Headers: drivers.NewHTTPHeaders(), StatusCode: 503}

	ordering, err := runtime.CompareValues(t.Context(), left, right)
	if err != nil {
		t.Fatalf("compare responses: %v", err)
	}
	if ordering != runtime.Less {
		t.Fatalf("response ordering = %d, want Less", ordering)
	}

	reverse, err := runtime.CompareValues(t.Context(), right, left)
	if err != nil {
		t.Fatalf("compare reversed responses: %v", err)
	}
	if reverse != runtime.Greater {
		t.Fatalf("reversed response ordering = %d, want Greater", reverse)
	}
}

func TestDriverComparisonsPropagateCancellationAndRejectIncompatibleValues(t *testing.T) {
	t.Parallel()

	headers := drivers.NewHTTPHeaders()
	canceled, cancel := context.WithCancel(t.Context())
	cancel()

	if _, err := headers.Compare(canceled, drivers.NewHTTPHeaders()); !errors.Is(err, context.Canceled) {
		t.Fatalf("Compare() error = %v, want context.Canceled", err)
	}
	if _, err := headers.Equal(canceled, drivers.NewHTTPHeaders()); !errors.Is(err, context.Canceled) {
		t.Fatalf("Equal() error = %v, want context.Canceled", err)
	}

	equal, err := runtime.EqualValues(t.Context(), headers, runtime.NewString("headers"))
	if err != nil {
		t.Fatalf("EqualValues() error = %v", err)
	}
	if equal {
		t.Fatal("incompatible values compared equal")
	}

	if _, err := runtime.CompareValues(t.Context(), headers, runtime.NewString("headers")); !errors.Is(err, runtime.ErrInvalidOperation) {
		t.Fatalf("CompareValues() error = %v, want ErrInvalidOperation", err)
	}
}

func TestMemoryAndCDPElementsRemainDistinctAndOrdered(t *testing.T) {
	t.Parallel()

	document, err := goquery.NewDocumentFromReader(strings.NewReader(`<div id="same">value</div>`))
	if err != nil {
		t.Fatalf("parse memory document: %v", err)
	}
	memoryElement, err := memory.NewHTMLElement(document, document.Find("#same"))
	if err != nil {
		t.Fatalf("create memory element: %v", err)
	}
	cdpElement := cdpdom.NewHTMLElement(
		zerolog.Nop(),
		nil,
		nil,
		nil,
		nil,
		cdpruntime.RemoteObjectID("same"),
	)
	matchingCDPElement := cdpdom.NewHTMLElement(
		zerolog.Nop(),
		nil,
		nil,
		nil,
		nil,
		cdpruntime.RemoteObjectID("same"),
	)
	assertEqualContract(t, cdpElement, matchingCDPElement)

	for _, operands := range [][2]runtime.Value{
		{memoryElement, cdpElement},
		{cdpElement, memoryElement},
	} {
		equal, err := runtime.EqualValues(t.Context(), operands[0], operands[1])
		if err != nil {
			t.Fatalf("EqualValues(%T, %T): %v", operands[0], operands[1], err)
		}
		if equal {
			t.Fatalf("cross-backend values %T and %T compared equal", operands[0], operands[1])
		}
	}

	ordering, err := runtime.CompareValues(t.Context(), memoryElement, cdpElement)
	if err != nil {
		t.Fatalf("compare memory to CDP: %v", err)
	}
	if ordering != runtime.Less {
		t.Fatalf("memory-to-CDP ordering = %d, want Less", ordering)
	}

	reverse, err := runtime.CompareValues(t.Context(), cdpElement, memoryElement)
	if err != nil {
		t.Fatalf("compare CDP to memory: %v", err)
	}
	if reverse != runtime.Greater {
		t.Fatalf("CDP-to-memory ordering = %d, want Greater", reverse)
	}
}

func TestMemoryDocumentNormalizedURLEqualityHasMatchingHash(t *testing.T) {
	t.Parallel()

	leftDocument, err := goquery.NewDocumentFromReader(strings.NewReader(`<html></html>`))
	if err != nil {
		t.Fatalf("parse left document: %v", err)
	}
	rightDocument, err := goquery.NewDocumentFromReader(strings.NewReader(`<html></html>`))
	if err != nil {
		t.Fatalf("parse right document: %v", err)
	}

	left, err := memory.NewRootHTMLDocument(leftDocument, "https://example.com/")
	if err != nil {
		t.Fatalf("create left document: %v", err)
	}
	right, err := memory.NewRootHTMLDocument(rightDocument, "https://example.com")
	if err != nil {
		t.Fatalf("create right document: %v", err)
	}

	assertEqualContract(t, left, right)
}

func assertEqualContract(t *testing.T, left, right runtime.Value) {
	t.Helper()

	equal, err := runtime.EqualValues(t.Context(), left, right)
	if err != nil {
		t.Fatalf("EqualValues(%T, %T): %v", left, right, err)
	}
	if !equal {
		t.Fatalf("EqualValues(%T, %T) = false", left, right)
	}

	ordering, err := runtime.CompareValues(t.Context(), left, right)
	if err != nil {
		t.Fatalf("CompareValues(%T, %T): %v", left, right, err)
	}
	if ordering != runtime.Equal {
		t.Fatalf("CompareValues(%T, %T) = %d, want Equal", left, right, ordering)
	}

	if left.Hash() != right.Hash() {
		t.Fatalf("equal values have different hashes: %d != %d", left.Hash(), right.Hash())
	}
}
