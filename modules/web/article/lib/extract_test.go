package lib

import (
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestExtractAndConvenienceFunctions(t *testing.T) {
	html := runtime.NewString(`
		<html lang="en">
		  <head>
		    <base href="https://example.com/articles/" />
		    <title>Ignored Title | Example News</title>
		    <meta property="og:title" content="OG Title" />
		    <meta property="og:description" content="OG description" />
		    <meta property="og:image" content="/images/hero.jpg" />
		    <link rel="canonical" href="/articles/structured" />
		    <script type="application/ld+json">
		      {
		        "@context": "https://schema.org",
		        "@type": "NewsArticle",
		        "headline": "Structured Title",
		        "description": "Structured description",
		        "author": { "@type": "Person", "name": "Jane Doe" },
		        "publisher": { "@type": "Organization", "name": "Example News" },
		        "datePublished": "2026-03-30T10:00:00Z",
		        "dateModified": "2026-03-30T12:00:00Z",
		        "image": "/images/hero.jpg",
		        "keywords": ["ai", "news"],
		        "articleSection": "Technology",
		        "url": "/articles/structured"
		      }
		    </script>
		  </head>
		  <body>
		    <article class="story article-content">
		      <header>
		        <h1>Structured Title</h1>
		        <p class="byline">By Jane Doe</p>
		      </header>
		      <p>Ferret article extraction focuses on the readable story body instead of navigation clutter. This paragraph carries enough substance to look article-like and should survive cleanup.</p>
		      <p>Additional body text keeps the candidate score high and ensures the text output is meaningful for downstream indexing and summarization workflows.</p>
		      <div class="share-tools">Share this story everywhere</div>
		    </article>
		  </body>
		</html>
	`)

	result, err := Extract(t.Context(), html)
	if err != nil {
		t.Fatalf("unexpected extract error: %v", err)
	}

	obj := mustRuntimeObject(t, result)

	if got := mustObjectField(t, t.Context(), obj, "title").String(); got != "Structured Title" {
		t.Fatalf("unexpected title %q", got)
	}

	if got := mustObjectField(t, t.Context(), obj, "byline").String(); got != "Jane Doe" {
		t.Fatalf("unexpected byline %q", got)
	}

	if got := mustObjectField(t, t.Context(), obj, "siteName").String(); got != "Example News" {
		t.Fatalf("unexpected siteName %q", got)
	}

	if got := mustObjectField(t, t.Context(), obj, "publishedAt").String(); got != "2026-03-30T10:00:00Z" {
		t.Fatalf("unexpected publishedAt %q", got)
	}

	if got := mustObjectField(t, t.Context(), obj, "canonicalUrl").String(); got != "https://example.com/articles/structured" {
		t.Fatalf("unexpected canonicalUrl %q", got)
	}

	if got := mustObjectField(t, t.Context(), obj, "leadImage").String(); got != "https://example.com/images/hero.jpg" {
		t.Fatalf("unexpected leadImage %q", got)
	}

	textValue := mustObjectField(t, t.Context(), obj, "text").String()
	if !strings.Contains(textValue, "readable story body") {
		t.Fatalf("expected article text, got %q", textValue)
	}

	if strings.Contains(textValue, "Share this story") {
		t.Fatalf("expected share block to be removed, got %q", textValue)
	}

	markdownValue, err := Markdown(t.Context(), html)
	if err != nil {
		t.Fatalf("unexpected markdown error: %v", err)
	}

	if markdownValue == runtime.None {
		t.Fatal("expected markdown output")
	}

	textOnly, err := Text(t.Context(), html)
	if err != nil {
		t.Fatalf("unexpected text error: %v", err)
	}

	if textOnly == runtime.None {
		t.Fatal("expected text output")
	}

	tags := mustArrayStrings(t, t.Context(), mustRuntimeArray(t, mustObjectField(t, t.Context(), obj, "tags")))
	if len(tags) != 2 || tags[0] != "ai" || tags[1] != "news" {
		t.Fatalf("unexpected tags %v", tags)
	}
}

func TestTextAndMarkdownReturnNoneForNonArticle(t *testing.T) {
	html := runtime.NewString(`
		<html>
		  <head>
		    <meta property="og:title" content="Portal Home" />
		  </head>
		  <body>
		    <nav>
		      <a href="/news">News</a>
		      <a href="/sports">Sports</a>
		      <a href="/weather">Weather</a>
		    </nav>
		  </body>
		</html>
	`)

	textValue, err := Text(t.Context(), html)
	if err != nil {
		t.Fatalf("unexpected text error: %v", err)
	}

	if textValue != runtime.None {
		t.Fatalf("expected runtime.None for text, got %v", textValue)
	}

	markdownValue, err := Markdown(t.Context(), html)
	if err != nil {
		t.Fatalf("unexpected markdown error: %v", err)
	}

	if markdownValue != runtime.None {
		t.Fatalf("expected runtime.None for markdown, got %v", markdownValue)
	}
}

func TestRejectsInvalidArgs(t *testing.T) {
	if _, err := Extract(t.Context()); err == nil {
		t.Fatal("expected error")
	}

	if _, err := Extract(t.Context(), runtime.NewInt(1)); err == nil {
		t.Fatal("expected type error")
	}

	if _, err := Text(t.Context(), runtime.NewString("a"), runtime.NewString("b")); err == nil {
		t.Fatal("expected arity error")
	}
}
