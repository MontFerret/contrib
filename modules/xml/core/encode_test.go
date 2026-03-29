package core

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type testClosableIterator struct {
	closeErr error
}

func (t *testClosableIterator) Next(context.Context) (runtime.Value, runtime.Value, error) {
	return runtime.None, runtime.None, io.EOF
}

func (t *testClosableIterator) Close() error {
	return t.closeErr
}

func TestEncodeCore(t *testing.T) {
	ctx := context.Background()

	t.Run("encodes document with sorted attrs and escaped text", func(t *testing.T) {
		value := xmlDocumentValue(xmlElementValue("book", map[string]runtime.Value{
			"z": runtime.NewString("1"),
			"a": runtime.NewString("2"),
		}, xmlTextValue("hi & <")))

		result, err := Encode(ctx, value)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != "<book a=\"2\" z=\"1\">hi &amp; &lt;</book>" {
			t.Fatalf("unexpected encode result: %q", result)
		}
	})

	t.Run("encodes element with missing attrs and children", func(t *testing.T) {
		value := runtime.NewObjectWith(map[string]runtime.Value{
			"type": runtime.NewString(nodeTypeElement),
			"name": runtime.NewString("book"),
		})

		result, err := Encode(ctx, value)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != "<book></book>" {
			t.Fatalf("unexpected encode result: %q", result)
		}
	})

	t.Run("rejects invalid top level node", func(t *testing.T) {
		_, err := Encode(ctx, xmlTextValue("hi"))
		if err == nil {
			t.Fatal("expected invalid top-level error")
		}
	})

	t.Run("rejects non element document roots", func(t *testing.T) {
		_, err := Encode(ctx, xmlDocumentValue(xmlTextValue("oops")))
		if err == nil {
			t.Fatal("expected document root validation error")
		}
	})

	t.Run("rejects document nodes inside children", func(t *testing.T) {
		value := xmlElementValue("root", nil, xmlDocumentValue(xmlElementValue("child", nil)))

		_, err := Encode(ctx, value)
		if err == nil {
			t.Fatal("expected nested document error")
		}
	})

	t.Run("rejects invalid attribute values and names", func(t *testing.T) {
		_, err := Encode(ctx, xmlElementValue("book", map[string]runtime.Value{
			"id": runtime.NewInt(123),
		}))
		if err == nil {
			t.Fatal("expected attribute type error")
		}

		_, err = Encode(ctx, xmlElementValue("bad name", nil))
		if err == nil {
			t.Fatal("expected invalid name error")
		}
	})
}

func TestEncodeHelpers(t *testing.T) {
	ctx := context.Background()

	t.Run("collectAttrs sorts names", func(t *testing.T) {
		attrs, err := collectAttrs(ctx, runtime.NewObjectWith(map[string]runtime.Value{
			"z": runtime.NewString("1"),
			"a": runtime.NewString("2"),
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(attrs) != 2 {
			t.Fatalf("expected 2 attrs, got %d", len(attrs))
		}

		if attrs[0].Name.Local != "a" || attrs[1].Name.Local != "z" {
			t.Fatalf("unexpected attr order: %#v", attrs)
		}
	})

	t.Run("field helper validation", func(t *testing.T) {
		node := runtime.NewObjectWith(map[string]runtime.Value{
			"type":     runtime.NewString(nodeTypeElement),
			"name":     runtime.NewString("book"),
			"attrs":    runtime.NewObjectWith(map[string]runtime.Value{"id": runtime.NewString("123")}),
			"children": runtime.NewArrayWith(xmlTextValue("hello")),
		})

		value, err := getRequiredField(ctx, node, "name")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertRuntimeStringValue(t, value, "book")

		name, err := getRequiredStringField(ctx, node, "name")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name != "book" {
			t.Fatalf("unexpected required string: %q", name)
		}

		attrs, err := getOptionalMapField(ctx, node, "attrs")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		length, err := attrs.Length(ctx)
		if err != nil {
			t.Fatalf("unexpected attrs length error: %v", err)
		}
		if length != 1 {
			t.Fatalf("expected attrs length 1, got %d", length)
		}

		children, err := getOptionalListField(ctx, node, "children")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		childLen, err := children.Length(ctx)
		if err != nil {
			t.Fatalf("unexpected children length error: %v", err)
		}
		if childLen != 1 {
			t.Fatalf("expected children length 1, got %d", childLen)
		}

		emptyNode := runtime.NewObject()
		attrs, err = getOptionalMapField(ctx, emptyNode, "attrs")
		if err != nil {
			t.Fatalf("unexpected empty attrs error: %v", err)
		}
		length, err = attrs.Length(ctx)
		if err != nil {
			t.Fatalf("unexpected empty attrs length error: %v", err)
		}
		if length != 0 {
			t.Fatalf("expected empty attrs length 0, got %d", length)
		}

		children, err = getOptionalListField(ctx, emptyNode, "children")
		if err != nil {
			t.Fatalf("unexpected empty children error: %v", err)
		}
		childLen, err = children.Length(ctx)
		if err != nil {
			t.Fatalf("unexpected empty children length error: %v", err)
		}
		if childLen != 0 {
			t.Fatalf("expected empty children length 0, got %d", childLen)
		}

		if _, err := getRequiredField(ctx, emptyNode, "missing"); err == nil {
			t.Fatal("expected missing field error")
		}

		if _, err := getRequiredStringField(ctx, runtime.NewObjectWith(map[string]runtime.Value{
			"name": runtime.NewInt(1),
		}), "name"); err == nil {
			t.Fatal("expected required string type error")
		}

		if _, err := getOptionalMapField(ctx, runtime.NewObjectWith(map[string]runtime.Value{
			"attrs": runtime.NewString("bad"),
		}), "attrs"); err == nil {
			t.Fatal("expected optional map type error")
		}

		if _, err := getOptionalListField(ctx, runtime.NewObjectWith(map[string]runtime.Value{
			"children": runtime.NewString("bad"),
		}), "children"); err == nil {
			t.Fatal("expected optional list type error")
		}

		if _, err := asNodeMap(runtime.NewString("bad")); err == nil {
			t.Fatal("expected node map type error")
		}
	})

	t.Run("closeIterator preserves existing errors and wraps close errors", func(t *testing.T) {
		var err error
		closeErr := errors.New("close boom")

		closeIterator(&testClosableIterator{closeErr: closeErr}, &err)
		if err == nil {
			t.Fatal("expected close error")
		}
		if !errors.Is(err, closeErr) {
			t.Fatalf("expected wrapped close error, got %v", err)
		}

		existing := errors.New("existing")
		err = existing
		closeIterator(&testClosableIterator{closeErr: closeErr}, &err)
		if !errors.Is(err, existing) {
			t.Fatalf("expected existing error to be preserved, got %v", err)
		}
	})
}
