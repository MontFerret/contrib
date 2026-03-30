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
		err := newError("boom")
		if err.Error() != "xml: boom" {
			t.Fatalf("unexpected error string: %q", err.Error())
		}
	})

	t.Run("wraps underlying errors", func(t *testing.T) {
		err := wrapError(io.EOF, "failed to decode")
		if err == nil {
			t.Fatal("expected wrapped error")
		}

		if !errors.Is(err, io.EOF) {
			t.Fatalf("expected wrapped error to unwrap to EOF, got %v", err)
		}

		if !strings.Contains(err.Error(), "xml: failed to decode") {
			t.Fatalf("unexpected wrapped error string: %q", err.Error())
		}

		xmlErr, ok := err.(*Error)
		if !ok {
			t.Fatalf("expected *XMLError, got %T", err)
		}

		if xmlErr.Unwrap() != io.EOF {
			t.Fatalf("unexpected unwrap result: %v", xmlErr.Unwrap())
		}
	})

	t.Run("ignores nil wrap input", func(t *testing.T) {
		if err := wrapError(nil, "ignored"); err != nil {
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

func TestNodeHelpers(t *testing.T) {
	ctx := context.Background()

	rootElement := xmlElementValue("book", map[string]runtime.Value{
		"id":    runtime.NewString("123"),
		"genre": runtime.NewString("fiction"),
	}, xmlTextValue("Hello "), xmlElementValue("title", nil, xmlTextValue("world")), xmlTextValue("!"))

	document := xmlDocumentValue(rootElement)
	textNode := xmlTextValue("standalone")

	t.Run("root unwraps documents and hides text nodes", func(t *testing.T) {
		result, err := Root(ctx, document)
		if err != nil {
			t.Fatalf("unexpected document root error: %v", err)
		}
		assertObjectFieldString(t, ctx, mustRuntimeObject(t, result), "name", "book")

		result, err = Root(ctx, rootElement)
		if err != nil {
			t.Fatalf("unexpected element root error: %v", err)
		}
		assertObjectFieldString(t, ctx, mustRuntimeObject(t, result), "name", "book")

		result, err = Root(ctx, textNode)
		if err != nil {
			t.Fatalf("unexpected text root error: %v", err)
		}
		if result != runtime.None {
			t.Fatalf("expected runtime.None, got %T", result)
		}
	})

	t.Run("text concatenates descendant text in document order", func(t *testing.T) {
		result, err := Text(ctx, document)
		if err != nil {
			t.Fatalf("unexpected document text error: %v", err)
		}
		assertRuntimeStringValue(t, result, "Hello world!")

		result, err = Text(ctx, rootElement)
		if err != nil {
			t.Fatalf("unexpected element text error: %v", err)
		}
		assertRuntimeStringValue(t, result, "Hello world!")

		result, err = Text(ctx, textNode)
		if err != nil {
			t.Fatalf("unexpected text-node text error: %v", err)
		}
		assertRuntimeStringValue(t, result, "standalone")

		emptyElement := xmlElementValue("empty", nil)
		result, err = Text(ctx, emptyElement)
		if err != nil {
			t.Fatalf("unexpected empty element text error: %v", err)
		}
		assertRuntimeStringValue(t, result, "")
	})

	t.Run("attr delegates through document roots and returns none when missing", func(t *testing.T) {
		result, err := Attr(ctx, document, runtime.NewString("id"))
		if err != nil {
			t.Fatalf("unexpected document attr error: %v", err)
		}
		assertRuntimeStringValue(t, result, "123")

		result, err = Attr(ctx, rootElement, runtime.NewString("genre"))
		if err != nil {
			t.Fatalf("unexpected element attr error: %v", err)
		}
		assertRuntimeStringValue(t, result, "fiction")

		result, err = Attr(ctx, rootElement, runtime.NewString("missing"))
		if err != nil {
			t.Fatalf("unexpected missing attr error: %v", err)
		}
		if result != runtime.None {
			t.Fatalf("expected runtime.None for missing attr, got %T", result)
		}

		result, err = Attr(ctx, textNode, runtime.NewString("missing"))
		if err != nil {
			t.Fatalf("unexpected text attr error: %v", err)
		}
		if result != runtime.None {
			t.Fatalf("expected runtime.None for text attr, got %T", result)
		}
	})

	t.Run("children delegate through document roots and return fresh empty arrays for text", func(t *testing.T) {
		result, err := Children(ctx, document)
		if err != nil {
			t.Fatalf("unexpected document children error: %v", err)
		}
		children := mustRuntimeArray(t, result)
		assertArrayLen(t, ctx, children, 3)
		assertObjectFieldString(t, ctx, mustArrayObjectAtIndex(t, ctx, children, 1), "name", "title")

		result, err = Children(ctx, rootElement)
		if err != nil {
			t.Fatalf("unexpected element children error: %v", err)
		}
		assertArrayLen(t, ctx, mustRuntimeArray(t, result), 3)

		firstTextChildren, err := Children(ctx, textNode)
		if err != nil {
			t.Fatalf("unexpected first text children error: %v", err)
		}
		secondTextChildren, err := Children(ctx, textNode)
		if err != nil {
			t.Fatalf("unexpected second text children error: %v", err)
		}

		firstArray := mustRuntimeArray(t, firstTextChildren)
		secondArray := mustRuntimeArray(t, secondTextChildren)
		assertArrayLen(t, ctx, firstArray, 0)
		assertArrayLen(t, ctx, secondArray, 0)
		if firstArray == secondArray {
			t.Fatal("expected text children helper to return a fresh array")
		}
	})

	t.Run("rejects non xml values and unsupported node types", func(t *testing.T) {
		if _, err := Root(ctx, runtime.NewInt(42)); err == nil || !errors.Is(err, runtime.ErrInvalidType) {
			t.Fatalf("expected runtime invalid type error, got %v", err)
		}

		invalidNode := runtime.NewObjectWith(map[string]runtime.Value{
			"type": runtime.NewString(eventTypeStartElement),
		})

		if _, err := Text(ctx, invalidNode); err == nil {
			t.Fatal("expected invalid node type error")
		} else {
			var xmlErr *Error
			if !errors.As(err, &xmlErr) {
				t.Fatalf("expected *XMLError, got %T", err)
			}
		}
	})

	t.Run("rejects malformed helper node shapes", func(t *testing.T) {
		if _, err := Root(ctx, xmlDocumentValue(xmlTextValue("oops"))); err == nil {
			t.Fatal("expected invalid document root error")
		}

		badAttrs := runtime.NewObjectWith(map[string]runtime.Value{
			"type":  runtime.NewString(nodeTypeElement),
			"name":  runtime.NewString("book"),
			"attrs": runtime.NewObjectWith(map[string]runtime.Value{"id": runtime.NewInt(123)}),
		})
		if _, err := Attr(ctx, badAttrs, runtime.NewString("id")); err == nil {
			t.Fatal("expected non-string attribute error")
		}

		nestedDocument := xmlElementValue("root", nil, xmlDocumentValue(xmlElementValue("child", nil)))
		if _, err := Text(ctx, nestedDocument); err == nil {
			t.Fatal("expected nested document helper error")
		}
	})
}
