package lib

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"

	htmldrivers "github.com/MontFerret/contrib/modules/web/html/drivers"
	htmlhttp "github.com/MontFerret/contrib/modules/web/html/drivers/memory"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestAcceptsHTMLModuleInputs(t *testing.T) {
	const pageHTML = `
		<html lang="en">
		  <head>
		    <title>Browser Rendered Story | Example Times</title>
		    <meta property="og:image" content="/media/hero.jpg" />
		    <link rel="canonical" href="/posts/browser-rendered-story" />
		  </head>
		  <body>
		    <article class="story article">
		      <h1>Browser Rendered Story</h1>
		      <img src="/media/hero-inline.jpg" alt="" />
		      <p>Browser-rendered content should still be extractable through the HTML module interfaces without forcing WEB::ARTICLE to understand drivers directly.</p>
		      <p>The second paragraph keeps the candidate container strong and makes sure the text output is meaningful for downstream usage.</p>
		    </article>
		  </body>
		</html>
	`

	page := mustHTMLPage(t, pageHTML, "https://example.com/posts/browser-rendered-story?ref=front")
	document := page.GetMainFrame()
	element := mustHTMLElement(t, pageHTML, "article")

	t.Run("page", func(t *testing.T) {
		result, err := Extract(t.Context(), page)
		if err != nil {
			t.Fatalf("unexpected extract error: %v", err)
		}

		obj := mustRuntimeObject(t, result)
		if got := mustObjectField(t, t.Context(), obj, "canonicalUrl").String(); got != "https://example.com/posts/browser-rendered-story" {
			t.Fatalf("unexpected canonicalUrl %q", got)
		}

		if got := mustObjectField(t, t.Context(), obj, "leadImage").String(); got != "https://example.com/media/hero.jpg" {
			t.Fatalf("unexpected leadImage %q", got)
		}

		textValue, err := Text(t.Context(), page)
		if err != nil {
			t.Fatalf("unexpected text error: %v", err)
		}

		if textValue == runtime.None || !strings.Contains(textValue.String(), "HTML module interfaces") {
			t.Fatalf("unexpected text output %v", textValue)
		}
	})

	t.Run("document", func(t *testing.T) {
		result, err := Extract(t.Context(), document)
		if err != nil {
			t.Fatalf("unexpected extract error: %v", err)
		}

		obj := mustRuntimeObject(t, result)
		if got := mustObjectField(t, t.Context(), obj, "siteName").String(); got != "Example Times" {
			t.Fatalf("unexpected siteName %q", got)
		}

		markdownValue, err := Markdown(t.Context(), document)
		if err != nil {
			t.Fatalf("unexpected markdown error: %v", err)
		}

		if markdownValue == runtime.None || !strings.Contains(markdownValue.String(), "Browser-rendered content should still be extractable") {
			t.Fatalf("unexpected markdown output %v", markdownValue)
		}
	})

	t.Run("element", func(t *testing.T) {
		result, err := Extract(t.Context(), element)
		if err != nil {
			t.Fatalf("unexpected extract error: %v", err)
		}

		obj := mustRuntimeObject(t, result)
		leadImage := mustObjectField(t, t.Context(), obj, "leadImage")
		if leadImage == runtime.None || leadImage.String() != "/media/hero-inline.jpg" {
			t.Fatalf("unexpected leadImage %v", leadImage)
		}

		if got := mustObjectField(t, t.Context(), obj, "canonicalUrl"); got != runtime.None {
			t.Fatalf("expected no canonicalUrl for element input, got %v", got)
		}

		if got := mustObjectField(t, t.Context(), obj, "title").String(); got != "Browser Rendered Story" {
			t.Fatalf("unexpected title %q", got)
		}
	})
}

func TestSerializeAttributesPreservesKeyValueOrderAndSkipsNone(t *testing.T) {
	attrs := runtime.NewObject()
	if err := attrs.Set(t.Context(), runtime.NewString("title"), runtime.NewString("hero image")); err != nil {
		t.Fatalf("unexpected set error: %v", err)
	}

	if err := attrs.Set(t.Context(), runtime.NewString("alt"), runtime.NewString("")); err != nil {
		t.Fatalf("unexpected set error: %v", err)
	}

	if err := attrs.Set(t.Context(), runtime.NewString("data-ignore"), runtime.None); err != nil {
		t.Fatalf("unexpected set error: %v", err)
	}

	if err := attrs.Set(t.Context(), runtime.NewString("class"), runtime.NewString("story hero")); err != nil {
		t.Fatalf("unexpected set error: %v", err)
	}

	serialized, err := serializeAttributes(t.Context(), attrs)
	if err != nil {
		t.Fatalf("unexpected serialize error: %v", err)
	}

	if serialized != ` alt="" class="story hero" title="hero image"` {
		t.Fatalf("unexpected serialized attrs %q", serialized)
	}
}

func TestSnapshotElementHTMLPreservesExistingAttributeNamesAndValues(t *testing.T) {
	element := mustHTMLElement(t, `<html><body><article class="story hero" data-slot="top story" title="Lead story"><p>Body</p></article></body></html>`, "article")

	htmlValue, err := snapshotElementHTML(t.Context(), element)
	if err != nil {
		t.Fatalf("unexpected snapshot error: %v", err)
	}

	if !strings.Contains(htmlValue, `<article class="story hero" data-slot="top story" title="Lead story">`) {
		t.Fatalf("unexpected snapshot html %q", htmlValue)
	}

	if strings.Contains(htmlValue, `story hero="class"`) || strings.Contains(htmlValue, `Lead story="title"`) || strings.Contains(htmlValue, `top story="data-slot"`) {
		t.Fatalf("expected attribute names and values to remain in the correct order, got %q", htmlValue)
	}
}

func mustHTMLPage(t *testing.T, input string, targetURL string) htmldrivers.HTMLPage {
	t.Helper()

	doc := mustGoqueryDocument(t, input)
	page, err := htmlhttp.NewHTMLPage(doc, targetURL, htmldrivers.HTTPResponse{
		URL:        targetURL,
		Status:     "200 OK",
		StatusCode: 200,
	}, htmldrivers.NewHTTPCookies())
	if err != nil {
		t.Fatalf("unexpected page error: %v", err)
	}

	return page
}

func mustHTMLElement(t *testing.T, input string, selector string) htmldrivers.HTMLElement {
	t.Helper()

	doc := mustGoqueryDocument(t, input)
	el, err := htmlhttp.NewHTMLElement(doc, doc.Find(selector).First())
	if err != nil {
		t.Fatalf("unexpected element error: %v", err)
	}

	return el
}

func mustGoqueryDocument(t *testing.T, input string) *goquery.Document {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected document parse error: %v", err)
	}

	return doc
}
