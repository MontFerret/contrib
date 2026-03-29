package core

import (
	"context"
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestXMLNameToString(t *testing.T) {
	t.Run("local name only", func(t *testing.T) {
		if got := xmlNameToString(xml.Name{Local: "book"}); got != "book" {
			t.Fatalf("unexpected name: %q", got)
		}
	})

	t.Run("qualified name", func(t *testing.T) {
		if got := xmlNameToString(xml.Name{Space: "ns", Local: "book"}); got != "ns:book" {
			t.Fatalf("unexpected name: %q", got)
		}
	})

	t.Run("space without local", func(t *testing.T) {
		if got := xmlNameToString(xml.Name{Space: "xmlns"}); got != "xmlns" {
			t.Fatalf("unexpected name: %q", got)
		}
	})
}

func TestXMLNameFromString(t *testing.T) {
	t.Run("preserves raw qualified names in local", func(t *testing.T) {
		name, err := xmlNameFromString("ns:book")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if name.Space != "" || name.Local != "ns:book" {
			t.Fatalf("unexpected xml.Name: %#v", name)
		}
	})

	t.Run("rejects empty names", func(t *testing.T) {
		_, err := xmlNameFromString("")
		if err == nil {
			t.Fatal("expected error for empty name")
		}
	})

	t.Run("rejects whitespace in names", func(t *testing.T) {
		_, err := xmlNameFromString("bad name")
		if err == nil {
			t.Fatal("expected error for whitespace in name")
		}
	})
}

func TestXMLErrorHelpers(t *testing.T) {
	t.Run("formats bare errors", func(t *testing.T) {
		err := newXMLError("boom")
		if err.Error() != "xml: boom" {
			t.Fatalf("unexpected error string: %q", err.Error())
		}
	})

	t.Run("wraps underlying errors", func(t *testing.T) {
		err := wrapXMLError(io.EOF, "failed to decode")
		if err == nil {
			t.Fatal("expected wrapped error")
		}

		if !errors.Is(err, io.EOF) {
			t.Fatalf("expected wrapped error to unwrap to EOF, got %v", err)
		}

		if !strings.Contains(err.Error(), "xml: failed to decode") {
			t.Fatalf("unexpected wrapped error string: %q", err.Error())
		}

		xmlErr, ok := err.(*XMLError)
		if !ok {
			t.Fatalf("expected *XMLError, got %T", err)
		}

		if xmlErr.Unwrap() != io.EOF {
			t.Fatalf("unexpected unwrap result: %v", xmlErr.Unwrap())
		}
	})

	t.Run("ignores nil wrap input", func(t *testing.T) {
		if err := wrapXMLError(nil, "ignored"); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestModelConstructors(t *testing.T) {
	ctx := context.Background()

	attrs := newAttrsObject([]xml.Attr{
		{Name: xml.Name{Local: "id"}, Value: "123"},
		{Name: xml.Name{Space: "xmlns", Local: "ns"}, Value: "urn:test"},
	})

	if got := attrs.String(); !strings.Contains(got, "\"id\":\"123\"") || !strings.Contains(got, "\"xmlns:ns\":\"urn:test\"") {
		t.Fatalf("unexpected attrs object: %s", got)
	}

	children := runtime.NewArray(0)
	element := newElementNode("ns:book", attrs, children)
	assertObjectFieldString(t, ctx, element, "type", nodeTypeElement)
	assertObjectFieldString(t, ctx, element, "name", "ns:book")
	assertObjectFieldString(t, ctx, mustObjectFieldObject(t, ctx, element, "attrs"), "id", "123")
	assertArrayLen(t, ctx, mustObjectFieldArray(t, ctx, element, "children"), 0)

	document := newDocumentNode(element)
	assertObjectFieldString(t, ctx, document, "type", nodeTypeDocument)
	root := mustObjectFieldObject(t, ctx, document, "root")
	assertObjectFieldString(t, ctx, root, "name", "ns:book")

	text := newTextNode("hello")
	assertObjectFieldString(t, ctx, text, "type", nodeTypeText)
	assertObjectFieldString(t, ctx, text, "value", "hello")

	start := newStartElementEvent("book", attrs)
	assertObjectFieldString(t, ctx, start, "type", eventTypeStartElement)
	assertObjectFieldString(t, ctx, start, "name", "book")

	end := newEndElementEvent("book")
	assertObjectFieldString(t, ctx, end, "type", eventTypeEndElement)
	assertObjectFieldString(t, ctx, end, "name", "book")
}
