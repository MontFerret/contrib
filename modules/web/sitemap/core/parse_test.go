package core

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	xmlcore "github.com/MontFerret/contrib/modules/xml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestParse(t *testing.T) {
	t.Run("parses urlset with namespaces and unknown tags", func(t *testing.T) {
		doc, err := Parse(t.Context(), strings.NewReader(`
			<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
			  <url>
			    <loc>https://example.com/a</loc>
			    <lastmod>2026-03-01T12:00:00Z</lastmod>
			    <changefreq>weekly</changefreq>
			    <priority>0.8</priority>
			    <image:image>
			      <image:loc>https://example.com/a.png</image:loc>
			    </image:image>
			  </url>
			  <url>
			    <loc>https://example.com/b</loc>
			  </url>
			</urlset>
		`), "https://example.com/sitemap.xml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if doc.Type != TypeURLSet {
			t.Fatalf("expected type %q, got %q", TypeURLSet, doc.Type)
		}

		if len(doc.URLs) != 2 {
			t.Fatalf("expected 2 URLs, got %d", len(doc.URLs))
		}

		first := doc.URLs[0]
		if first.Loc != "https://example.com/a" {
			t.Fatalf("unexpected first loc %q", first.Loc)
		}

		if first.LastMod != "2026-03-01T12:00:00Z" {
			t.Fatalf("unexpected lastmod %q", first.LastMod)
		}

		if first.ChangeFreq != "weekly" {
			t.Fatalf("unexpected changefreq %q", first.ChangeFreq)
		}

		if first.Priority == nil || *first.Priority != 0.8 {
			t.Fatalf("expected priority 0.8, got %v", first.Priority)
		}

		if first.Source != "https://example.com/sitemap.xml" {
			t.Fatalf("unexpected source %q", first.Source)
		}

		second := doc.URLs[1]
		if second.Loc != "https://example.com/b" {
			t.Fatalf("unexpected second loc %q", second.Loc)
		}

		if second.LastMod != "" || second.ChangeFreq != "" || second.Priority != nil {
			t.Fatalf("expected optional fields to be empty, got %+v", second)
		}
	})

	t.Run("parses sitemapindex with prefixed namespaces", func(t *testing.T) {
		doc, err := Parse(t.Context(), strings.NewReader(`
			<sm:sitemapindex xmlns:sm="http://www.sitemaps.org/schemas/sitemap/0.9">
			  <sm:sitemap>
			    <sm:loc>https://example.com/posts.xml</sm:loc>
			    <sm:lastmod>2026-03-02</sm:lastmod>
			  </sm:sitemap>
			</sm:sitemapindex>
		`), "https://example.com/index.xml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if doc.Type != TypeSitemapIndex {
			t.Fatalf("expected type %q, got %q", TypeSitemapIndex, doc.Type)
		}

		if len(doc.Sitemaps) != 1 {
			t.Fatalf("expected 1 sitemap, got %d", len(doc.Sitemaps))
		}

		ref := doc.Sitemaps[0]
		if ref.Loc != "https://example.com/posts.xml" {
			t.Fatalf("unexpected loc %q", ref.Loc)
		}

		if ref.LastMod != "2026-03-02" {
			t.Fatalf("unexpected lastmod %q", ref.LastMod)
		}
	})

	t.Run("rejects unsupported roots", func(t *testing.T) {
		_, err := Parse(t.Context(), strings.NewReader(`<feed></feed>`), "https://example.com/feed.xml")
		if err == nil {
			t.Fatal("expected error")
		}

		assertStageError(t, err, StageParse, "https://example.com/feed.xml")
	})

	t.Run("rejects malformed xml", func(t *testing.T) {
		_, err := Parse(t.Context(), strings.NewReader(`<urlset><url><loc>https://example.com</url></urlset>`), "https://example.com/bad.xml")
		if err == nil {
			t.Fatal("expected error")
		}

		assertStageError(t, err, StageParse, "https://example.com/bad.xml")
	})

	t.Run("returns parse-stage cancellation when context is already canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		_, err := Parse(ctx, strings.NewReader(`<urlset/>`), "https://example.com/canceled.xml")
		if err == nil {
			t.Fatal("expected error")
		}

		assertStageError(t, err, StageParse, "https://example.com/canceled.xml")
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected wrapped context.Canceled, got %v", err)
		}
	})

	t.Run("xml core iterator remains compatible with sitemap interpreter", func(t *testing.T) {
		iter, err := xmlcore.NewDecodeIteratorFromReader(strings.NewReader(`
			<sm:sitemapindex xmlns:sm="http://www.sitemaps.org/schemas/sitemap/0.9">
			  <sm:sitemap>
			    <sm:loc>https://example.com/posts.xml</sm:loc>
			  </sm:sitemap>
			</sm:sitemapindex>
		`))
		if err != nil {
			t.Fatalf("unexpected iterator error: %v", err)
		}

		doc, err := parseIterator(t.Context(), iter, "https://example.com/index.xml")
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}

		if doc.Type != TypeSitemapIndex {
			t.Fatalf("expected type %q, got %q", TypeSitemapIndex, doc.Type)
		}

		if len(doc.Sitemaps) != 1 || doc.Sitemaps[0].Loc != "https://example.com/posts.xml" {
			t.Fatalf("unexpected sitemaps %+v", doc.Sitemaps)
		}
	})

	t.Run("concatenates split text events and preserves sibling fields", func(t *testing.T) {
		iter := &eventIterator{
			events: []runtime.Value{
				xmlStartEvent(TypeURLSet),
				xmlStartEvent("url"),
				xmlStartEvent("loc"),
				xmlTextEvent("https://example.com/"),
				xmlTextEvent("split"),
				xmlEndEvent("loc"),
				xmlStartEvent("lastmod"),
				xmlTextEvent("2026-03-01"),
				xmlEndEvent("lastmod"),
				xmlEndEvent("url"),
				xmlEndEvent(TypeURLSet),
			},
		}

		doc, err := parseIterator(t.Context(), iter, "https://example.com/split.xml")
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}

		if doc.Type != TypeURLSet {
			t.Fatalf("expected type %q, got %q", TypeURLSet, doc.Type)
		}

		if len(doc.URLs) != 1 {
			t.Fatalf("expected 1 URL, got %d", len(doc.URLs))
		}

		entry := doc.URLs[0]
		if entry.Loc != "https://example.com/split" {
			t.Fatalf("unexpected loc %q", entry.Loc)
		}

		if entry.LastMod != "2026-03-01" {
			t.Fatalf("unexpected lastmod %q", entry.LastMod)
		}
	})

	t.Run("parse iterator stops on context cancellation before consuming another event", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		iter := newBlockingIterator()

		result := make(chan error, 1)
		go func() {
			_, err := parseIterator(ctx, iter, "https://example.com/blocked.xml")
			result <- err
		}()

		<-iter.firstCalled
		cancel()
		close(iter.releaseFirst)

		select {
		case err := <-result:
			if err == nil {
				t.Fatal("expected error")
			}

			assertStageError(t, err, StageParse, "https://example.com/blocked.xml")
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("expected wrapped context.Canceled, got %v", err)
			}
		case <-time.After(200 * time.Millisecond):
			t.Fatal("parseIterator did not stop after cancellation")
		}

		select {
		case <-iter.secondCalled:
			t.Fatal("expected parseIterator to stop before requesting another event")
		default:
		}
	})
}

type blockingIterator struct {
	firstCalled  chan struct{}
	secondCalled chan struct{}
	releaseFirst chan struct{}
	calls        int
}

type eventIterator struct {
	events []runtime.Value
	index  int
}

func newBlockingIterator() *blockingIterator {
	return &blockingIterator{
		firstCalled:  make(chan struct{}),
		secondCalled: make(chan struct{}),
		releaseFirst: make(chan struct{}),
	}
}

func (i *blockingIterator) Next(_ context.Context) (runtime.Value, runtime.Value, error) {
	i.calls++

	if i.calls == 1 {
		close(i.firstCalled)
		<-i.releaseFirst

		return runtime.NewObjectWith(map[string]runtime.Value{
			"type": runtime.NewString(xmlEventStartElement),
			"name": runtime.NewString(TypeURLSet),
		}), runtime.NewInt(1), nil
	}

	select {
	case <-i.secondCalled:
	default:
		close(i.secondCalled)
	}

	return runtime.None, runtime.None, io.EOF
}

func (i *eventIterator) Next(_ context.Context) (runtime.Value, runtime.Value, error) {
	if i.index >= len(i.events) {
		return runtime.None, runtime.None, io.EOF
	}

	value := i.events[i.index]
	i.index++

	return value, runtime.NewInt(i.index), nil
}

func xmlStartEvent(name string) runtime.Value {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"type": runtime.NewString(xmlEventStartElement),
		"name": runtime.NewString(name),
	})
}

func xmlEndEvent(name string) runtime.Value {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"type": runtime.NewString(xmlEventEndElement),
		"name": runtime.NewString(name),
	})
}

func xmlTextEvent(value string) runtime.Value {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"type":  runtime.NewString(xmlEventText),
		"value": runtime.NewString(value),
	})
}
