package lib

import (
	"context"
	"testing"

	"github.com/MontFerret/contrib/modules/xml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestEncodeLib(t *testing.T) {
	ctx := context.Background()

	t.Run("encodes element with sorted attrs and escaped text", func(t *testing.T) {
		value := xmlElement("book", map[string]runtime.Value{
			"z": runtime.NewString("1"),
			"a": runtime.NewString("2"),
		}, xmlText("hi & <"))

		result, err := Encode(ctx, value)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.String() != "<book a=\"2\" z=\"1\">hi &amp; &lt;</book>" {
			t.Fatalf("unexpected encode result: %q", result.String())
		}
	})

	t.Run("encodes document and preserves qualified names", func(t *testing.T) {
		value := xmlDocument(xmlElement("ns:book", map[string]runtime.Value{
			"xmlns:ns": runtime.NewString("urn:test"),
			"id":       runtime.NewString("123"),
		}, xmlElement("ns:title", nil, xmlText("Hello"))))

		result, err := Encode(ctx, value)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := "<ns:book id=\"123\" xmlns:ns=\"urn:test\"><ns:title>Hello</ns:title></ns:book>"
		if result.String() != expected {
			t.Fatalf("unexpected encode result: got %q want %q", result.String(), expected)
		}
	})

	t.Run("defaults missing attrs and children to empty", func(t *testing.T) {
		value := runtime.NewObjectWith(map[string]runtime.Value{
			"type": runtime.NewString("element"),
			"name": runtime.NewString("book"),
		})

		result, err := Encode(ctx, value)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.String() != "<book></book>" {
			t.Fatalf("unexpected encode result: %q", result.String())
		}
	})

	t.Run("rejects invalid top level nodes", func(t *testing.T) {
		_, err := Encode(ctx, xmlText("hi"))
		if err == nil {
			t.Fatal("expected error for top-level text node")
		}
	})

	t.Run("rejects non string attributes", func(t *testing.T) {
		value := xmlElement("book", map[string]runtime.Value{
			"id": runtime.NewInt(123),
		})

		_, err := Encode(ctx, value)
		if err == nil {
			t.Fatal("expected attribute validation error")
		}

		if _, ok := err.(*core.XMLError); !ok {
			t.Fatalf("expected *core.XMLError, got %T", err)
		}
	})

	t.Run("round trips decode encode decode", func(t *testing.T) {
		decoded, err := Decode(ctx, runtime.NewString("<ns:book xmlns:ns=\"urn:test\"><ns:title>Hello</ns:title><body>Hi there</body></ns:book>"))
		if err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}

		encoded, err := Encode(ctx, decoded)
		if err != nil {
			t.Fatalf("unexpected encode error: %v", err)
		}

		decodedAgain, err := Decode(ctx, encoded)
		if err != nil {
			t.Fatalf("unexpected round-trip decode error: %v", err)
		}

		if decoded.String() != decodedAgain.String() {
			t.Fatalf("round-trip mismatch: got %s want %s", decodedAgain.String(), decoded.String())
		}
	})
}
