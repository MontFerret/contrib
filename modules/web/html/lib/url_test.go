package lib

import (
	"context"
	"errors"
	"testing"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestURLFunctionsUseDocumentURLCapability(t *testing.T) {
	t.Parallel()

	page := newMemoryPage(
		t,
		`<html><head><base href="/assets/"></head><body></body></html>`,
		nil,
	)
	doc := page.GetMainFrame()
	ctx := context.Background()

	for _, root := range []runtime.Value{page, doc} {
		got, err := URL(ctx, root)
		if err != nil {
			t.Fatalf("current URL: %v", err)
		}
		if got.String() != "https://example.com" {
			t.Fatalf("expected current URL, got %q", got)
		}

		got, err = BaseURL(ctx, root)
		if err != nil {
			t.Fatalf("base URL: %v", err)
		}
		if got.String() != "https://example.com/assets/" {
			t.Fatalf("expected base URL, got %q", got)
		}

		got, err = ResolveURL(ctx, root, runtime.NewString("image.png"))
		if err != nil {
			t.Fatalf("resolve URL: %v", err)
		}
		if got.String() != "https://example.com/assets/image.png" {
			t.Fatalf("expected resolved URL, got %q", got)
		}
	}
}

func TestURLFunctionsValidateArguments(t *testing.T) {
	t.Parallel()

	page := newMemoryPage(t, `<html><body></body></html>`, nil)
	ctx := context.Background()

	cases := []struct {
		call func() (runtime.Value, error)
		name string
	}{
		{name: "URL invalid root", call: func() (runtime.Value, error) {
			return URL(ctx, runtime.NewString("https://example.com"))
		}},
		{name: "BASE_URL invalid root", call: func() (runtime.Value, error) {
			return BaseURL(ctx, runtime.NewObject())
		}},
		{name: "RESOLVE_URL invalid root", call: func() (runtime.Value, error) {
			return ResolveURL(ctx, runtime.NewString("https://example.com"), runtime.NewString("asset.png"))
		}},
		{name: "RESOLVE_URL invalid URL type", call: func() (runtime.Value, error) {
			return ResolveURL(ctx, page, runtime.NewInt(1))
		}},
		{name: "RESOLVE_URL malformed URL", call: func() (runtime.Value, error) {
			return ResolveURL(ctx, page, runtime.NewString("%zz"))
		}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := tc.call(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestURLFunctionsRejectUnsupportedDocumentCapability(t *testing.T) {
	t.Parallel()

	doc := &documentWithoutURLCapability{
		HTMLDocument: newMemoryDocument(t, `<html><body></body></html>`),
	}

	_, err := URL(context.Background(), doc)
	if !errors.Is(err, runtime.ErrNotSupported) {
		t.Fatalf("expected unsupported URL capability error, got %v", err)
	}
}

type documentWithoutURLCapability struct {
	drivers.HTMLDocument
}
