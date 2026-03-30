package core

import (
	"net/url"
	"strings"
	"testing"
)

func TestExtractStructuredArticle(t *testing.T) {
	article := Extract(`
		<html lang="en" dir="ltr">
		  <head>
		    <base href="https://example.com/posts/" />
		    <title>Ignored Window Title | Example News</title>
		    <meta property="og:title" content="OG Title" />
		    <meta property="og:description" content="OG description" />
		    <meta property="og:image" content="/images/lead.jpg" />
		    <meta property="og:site_name" content="Example News" />
		    <link rel="canonical" href="/posts/structured-title" />
		    <script type="application/ld+json">
		      {
		        "@context": "https://schema.org",
		        "@graph": [
		          {
		            "@type": "NewsArticle",
		            "headline": "Structured Title",
		            "description": "Structured summary from JSON-LD",
		            "author": { "@type": "Person", "name": "Jane Doe" },
		            "publisher": { "@type": "Organization", "name": "Example News" },
		            "datePublished": "2026-03-30T10:00:00Z",
		            "dateModified": "2026-03-30T12:00:00Z",
		            "image": "/images/lead.jpg",
		            "keywords": ["ai", "news"],
		            "articleSection": "Technology",
		            "url": "/posts/structured-title"
		          }
		        ]
		      }
		    </script>
		  </head>
		  <body>
		    <article class="story article-content">
		      <header class="article-header">
		        <h1>Structured Title</h1>
		        <p class="byline">By Jane Doe</p>
		        <time datetime="2026-03-30T10:00:00Z">March 30, 2026</time>
		      </header>
		      <p>Ferret article extraction focuses on the readable story body instead of navigation clutter. This opening paragraph should survive cleanup and remain in the final text output for indexing pipelines.</p>
		      <p>The second paragraph adds more signal and keeps the candidate container strong enough to beat unrelated sections such as share tools, recommended content, and comment rails.</p>
		      <div class="share-tools">Share this story</div>
		      <section class="related-posts">
		        <a href="/posts/other">Other story</a>
		      </section>
		      <section class="comments">Comment thread here</section>
		    </article>
		  </body>
		</html>
	`)

	if article.Title == nil || *article.Title != "Structured Title" {
		t.Fatalf("unexpected title %+v", article.Title)
	}

	if article.Byline == nil || *article.Byline != "Jane Doe" {
		t.Fatalf("unexpected byline %+v", article.Byline)
	}

	if article.Excerpt == nil || *article.Excerpt != "Structured summary from JSON-LD" {
		t.Fatalf("unexpected excerpt %+v", article.Excerpt)
	}

	if article.SiteName == nil || *article.SiteName != "Example News" {
		t.Fatalf("unexpected siteName %+v", article.SiteName)
	}

	if article.PublishedAt == nil || *article.PublishedAt != "2026-03-30T10:00:00Z" {
		t.Fatalf("unexpected publishedAt %+v", article.PublishedAt)
	}

	if article.UpdatedAt == nil || *article.UpdatedAt != "2026-03-30T12:00:00Z" {
		t.Fatalf("unexpected updatedAt %+v", article.UpdatedAt)
	}

	if article.CanonicalURL == nil || *article.CanonicalURL != "https://example.com/posts/structured-title" {
		t.Fatalf("unexpected canonical url %+v", article.CanonicalURL)
	}

	if article.LeadImage == nil || *article.LeadImage != "https://example.com/images/lead.jpg" {
		t.Fatalf("unexpected lead image %+v", article.LeadImage)
	}

	if article.Text == nil || !strings.Contains(*article.Text, "readable story body") {
		t.Fatalf("unexpected text %+v", article.Text)
	}

	if strings.Contains(*article.Text, "Share this story") || strings.Contains(*article.Text, "Comment thread") {
		t.Fatalf("expected cleanup to remove boilerplate, got %q", *article.Text)
	}

	if article.HTML == nil || strings.Contains(*article.HTML, "related-posts") || strings.Contains(*article.HTML, "share-tools") {
		t.Fatalf("unexpected cleaned html %+v", article.HTML)
	}

	if article.Markdown == nil || strings.Contains(*article.Markdown, "Structured Title") {
		t.Fatalf("unexpected markdown %+v", article.Markdown)
	}

	if article.WordCount == nil || *article.WordCount < 35 {
		t.Fatalf("unexpected wordCount %+v", article.WordCount)
	}

	if article.ReadingTimeMinutes == nil || *article.ReadingTimeMinutes != 1 {
		t.Fatalf("unexpected readingTimeMinutes %+v", article.ReadingTimeMinutes)
	}

	if len(article.Tags) != 2 || article.Tags[0] != "ai" || article.Tags[1] != "news" {
		t.Fatalf("unexpected tags %v", article.Tags)
	}

	if len(article.Categories) != 1 || article.Categories[0] != "Technology" {
		t.Fatalf("unexpected categories %v", article.Categories)
	}
}

func TestExtractDocsLikePage(t *testing.T) {
	article := Extract(`
		<html lang="en">
		  <head>
		    <title>API Guide | Example Docs</title>
		  </head>
		  <body>
		    <main id="docs-content" class="main docs-content">
		      <nav class="sidebar-nav"><a href="/docs/start">Start</a><a href="/docs/auth">Auth</a></nav>
		      <section>
		        <h1>API Guide</h1>
		        <p>This guide explains how to call the Example API, authenticate requests, and interpret the response body for automation and indexing workflows.</p>
		        <pre><code>GET /v1/items
Authorization: Bearer token</code></pre>
		        <table>
		          <tr><td>Field</td><td>Type</td></tr>
		          <tr><td>id</td><td>string</td></tr>
		          <tr><td>name</td><td>string</td></tr>
		        </table>
		        <p>Use the endpoint responsibly and cache responses when possible to avoid unnecessary load.</p>
		      </section>
		    </main>
		  </body>
		</html>
	`)

	if article.Title == nil || *article.Title != "API Guide" {
		t.Fatalf("unexpected title %+v", article.Title)
	}

	if article.SiteName == nil || *article.SiteName != "Example Docs" {
		t.Fatalf("unexpected siteName %+v", article.SiteName)
	}

	if article.Text == nil || !strings.Contains(*article.Text, "GET /v1/items") {
		t.Fatalf("unexpected text %+v", article.Text)
	}

	if strings.Contains(*article.Text, "sidebar-nav") || strings.Contains(*article.Text, "Start Auth") {
		t.Fatalf("expected sidebar cleanup, got %q", *article.Text)
	}

	if article.Markdown == nil || !strings.Contains(*article.Markdown, "| Field") || !strings.Contains(*article.Markdown, "GET /v1/items") {
		t.Fatalf("unexpected markdown %+v", article.Markdown)
	}
}

func TestExtractMetadataOnlyPage(t *testing.T) {
	article := Extract(`
		<html>
		  <head>
		    <meta property="og:title" content="Portal Home" />
		    <meta property="og:description" content="Landing page description" />
		    <meta property="article:modified_time" content="Spring 2026" />
		    <link rel="canonical" href="/portal" />
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

	if article.Title == nil || *article.Title != "Portal Home" {
		t.Fatalf("unexpected title %+v", article.Title)
	}

	if article.Excerpt == nil || *article.Excerpt != "Landing page description" {
		t.Fatalf("unexpected excerpt %+v", article.Excerpt)
	}

	if article.UpdatedAt == nil || *article.UpdatedAt != "Spring 2026" {
		t.Fatalf("unexpected updatedAt %+v", article.UpdatedAt)
	}

	if article.CanonicalURL == nil || *article.CanonicalURL != "/portal" {
		t.Fatalf("unexpected canonical url %+v", article.CanonicalURL)
	}

	if article.Text != nil || article.HTML != nil || article.Markdown != nil {
		t.Fatalf("expected nil body fields, got %+v", article)
	}
}

func TestExtractMalformedHTML(t *testing.T) {
	article := Extract(`
		<html>
		  <head><title>Broken Example | Example Site</title></head>
		  <body>
		    <article class="content">
		      <h1>Broken Example</h1>
		      <p>Broken markup still contains enough readable prose to form an article extraction result for best-effort parsing, even when tags are not properly balanced and the tree needs recovery.
		      <p>The parser should recover and keep the article body accessible for downstream processing, indexing, and markdown conversion without treating the page as empty boilerplate.
		    </article>
		  </body>
		</html>
	`)

	if article.Title == nil || *article.Title != "Broken Example" {
		t.Fatalf("unexpected title %+v", article.Title)
	}

	if article.Text == nil || !strings.Contains(*article.Text, "best-effort parsing") {
		t.Fatalf("unexpected text %+v", article.Text)
	}
}

func TestExtractSourceUsesSourceURLFallback(t *testing.T) {
	sourceURL, err := url.Parse("https://example.com/docs/rendered-guide")
	if err != nil {
		t.Fatalf("unexpected url parse error: %v", err)
	}

	article := ExtractSource(Source{
		HTML: `
			<html>
			  <head>
			    <link rel="canonical" href="/docs/rendered-guide" />
			    <meta property="og:image" content="/media/guide.jpg" />
			  </head>
			  <body>
			    <article class="guide">
			      <p>This rendered guide is extracted from a DOM snapshot and should still resolve relative metadata URLs by falling back to the source page URL.</p>
			      <p>That fallback matters for pages opened through the HTML module where there is no explicit base tag in the document.</p>
			    </article>
			  </body>
			</html>
		`,
		SourceURL: sourceURL,
	})

	if article.CanonicalURL == nil || *article.CanonicalURL != "https://example.com/docs/rendered-guide" {
		t.Fatalf("unexpected canonicalUrl %+v", article.CanonicalURL)
	}

	if article.LeadImage == nil || *article.LeadImage != "https://example.com/media/guide.jpg" {
		t.Fatalf("unexpected leadImage %+v", article.LeadImage)
	}
}

func TestExtractSourceUsesTitleHintAsFallback(t *testing.T) {
	article := ExtractSource(Source{
		HTML: `
			<html>
			  <body>
			    <main class="docs-content">
			      <h1>Rendered Guide</h1>
			      <p>This guide was loaded through a rendered DOM snapshot and still needs sensible title and site-name fallback behavior.</p>
			      <p>The document lacks a title element, so the source title hint should be able to fill the publication name without overriding the in-body heading.</p>
			    </main>
			  </body>
			</html>
		`,
		TitleHint: stringPtr("Rendered Guide | Example Docs"),
	})

	if article.Title == nil || *article.Title != "Rendered Guide" {
		t.Fatalf("unexpected title %+v", article.Title)
	}

	if article.SiteName == nil || *article.SiteName != "Example Docs" {
		t.Fatalf("unexpected siteName %+v", article.SiteName)
	}
}
