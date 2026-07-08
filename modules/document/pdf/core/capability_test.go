package core

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDocumentAndPageCapabilities(t *testing.T) {
	t.Parallel()

	ctx, root := testFSContext(t, false)
	writePDFForTest(t, root, "capabilities.pdf", []pdfTestPage{
		{Text: "Alpha", Width: 612, Height: 792, Rotation: 90},
		{Text: "Beta", Width: 400, Height: 500, X: 36, Y: 420, FontSize: 18},
	})

	document, err := Open(ctx, "capabilities.pdf")
	if err != nil {
		t.Fatalf("unexpected open error: %v", err)
	}
	t.Cleanup(func() { _ = document.Close() })

	countValue, err := document.Get(ctx, runtime.NewString("pageCount"))
	if err != nil {
		t.Fatalf("unexpected pageCount property error: %v", err)
	}
	if countValue != runtime.NewInt(2) {
		t.Fatalf("pageCount = %v, want 2", countValue)
	}

	pagesValue, err := document.Get(ctx, runtime.NewString("pages"))
	if err != nil {
		t.Fatalf("unexpected pages property error: %v", err)
	}
	pages, ok := pagesValue.(*PageCollection)
	if !ok {
		t.Fatalf("expected PageCollection, got %T", pagesValue)
	}
	if pages.String() != "<document.pdf.pages>" || pages.Hash() == 0 || pages.Copy() != pages {
		t.Fatal("expected page collection to expose stable runtime value identity")
	}

	length, err := pages.Length(ctx)
	if err != nil {
		t.Fatalf("unexpected length error: %v", err)
	}
	if length != runtime.NewInt(2) {
		t.Fatalf("length = %v, want 2", length)
	}

	firstValue, err := pages.At(ctx, runtime.ZeroInt)
	if err != nil {
		t.Fatalf("unexpected first page access error: %v", err)
	}
	first, ok := firstValue.(*Page)
	if !ok {
		t.Fatalf("expected first page, got %T", firstValue)
	}
	if first.Number() != 1 {
		t.Fatalf("first page number = %d, want 1", first.Number())
	}

	secondValue, found, err := pages.LookupAt(ctx, runtime.NewInt(1))
	if err != nil {
		t.Fatalf("unexpected second page lookup error: %v", err)
	}
	if !found {
		t.Fatal("expected second page lookup to be found")
	}
	second := secondValue.(*Page)
	if second.Number() != 2 {
		t.Fatalf("second page number = %d, want 2", second.Number())
	}

	if value, err := pages.At(ctx, runtime.NewInt(-1)); err != nil || value != runtime.None {
		t.Fatalf("negative index = %v, %v; want NONE, nil", value, err)
	}
	if value, found, err := pages.LookupAt(ctx, runtime.NewInt(99)); err != nil || found || value != runtime.None {
		t.Fatalf("out-of-range lookup = %v, %v, %v; want NONE, false, nil", value, found, err)
	}

	number, err := first.Get(ctx, runtime.NewString("number"))
	if err != nil {
		t.Fatalf("unexpected number property error: %v", err)
	}
	if number != runtime.NewInt(1) {
		t.Fatalf("number = %v, want 1", number)
	}
	width, err := first.Get(ctx, runtime.NewString("width"))
	if err != nil {
		t.Fatalf("unexpected width property error: %v", err)
	}
	if width != runtime.NewFloat(612) {
		t.Fatalf("width = %v, want 612", width)
	}
	height, err := first.Get(ctx, runtime.NewString("height"))
	if err != nil {
		t.Fatalf("unexpected height property error: %v", err)
	}
	if height != runtime.NewFloat(792) {
		t.Fatalf("height = %v, want 792", height)
	}
	rotation, err := first.Get(ctx, runtime.NewString("rotation"))
	if err != nil {
		t.Fatalf("unexpected rotation property error: %v", err)
	}
	if rotation != runtime.NewInt(90) {
		t.Fatalf("rotation = %v, want 90", rotation)
	}

	text, err := first.Get(ctx, runtime.NewString("text"))
	if err != nil {
		t.Fatalf("unexpected text property error: %v", err)
	}
	if !strings.Contains(text.String(), "Alpha") {
		t.Fatalf("text property = %q, want fixture text", text.String())
	}

	blocks, err := first.Get(ctx, runtime.NewString("blocks"))
	if err != nil {
		t.Fatalf("unexpected blocks property error: %v", err)
	}
	list, ok := blocks.(runtime.List)
	if !ok {
		t.Fatalf("expected blocks list, got %T", blocks)
	}
	if length, err := list.Length(ctx); err != nil || length == 0 {
		t.Fatalf("blocks length = %v, %v; want non-zero", length, err)
	}

	if value, err := document.Get(ctx, runtime.NewString("missing")); err != nil || value != runtime.None {
		t.Fatalf("unknown document key = %v, %v; want NONE, nil", value, err)
	}
	if value, err := first.Get(ctx, runtime.NewString("missing")); err != nil || value != runtime.None {
		t.Fatalf("unknown page key = %v, %v; want NONE, nil", value, err)
	}
	if value, err := document.Get(ctx, runtime.None); err != nil || value != runtime.None {
		t.Fatalf("empty document key = %v, %v; want NONE, nil", value, err)
	}
	if value, err := first.Get(ctx, runtime.None); err != nil || value != runtime.None {
		t.Fatalf("empty page key = %v, %v; want NONE, nil", value, err)
	}
}

func TestPageCollectionIteratesLazily(t *testing.T) {
	t.Parallel()

	ctx, root := testFSContext(t, false)
	writePDFForTest(t, root, "iter.pdf", []pdfTestPage{
		{Text: "One"},
		{Text: "Two"},
	})

	document, err := Open(ctx, "iter.pdf")
	if err != nil {
		t.Fatalf("unexpected open error: %v", err)
	}
	t.Cleanup(func() { _ = document.Close() })

	pages := NewPageCollection(document)
	iter, err := pages.Iterate(ctx)
	if err != nil {
		t.Fatalf("unexpected iterate error: %v", err)
	}

	first, key, err := iter.Next(ctx)
	if err != nil {
		t.Fatalf("unexpected first next error: %v", err)
	}
	if key != runtime.ZeroInt || first.(*Page).Number() != 1 {
		t.Fatalf("unexpected first iteration result: value=%v key=%v", first, key)
	}

	second, key, err := iter.Next(ctx)
	if err != nil {
		t.Fatalf("unexpected second next error: %v", err)
	}
	if key != runtime.NewInt(1) || second.(*Page).Number() != 2 {
		t.Fatalf("unexpected second iteration result: value=%v key=%v", second, key)
	}

	if value, key, err := iter.Next(ctx); !errors.Is(err, io.EOF) || value != runtime.None || key != runtime.None {
		t.Fatalf("final next = %v, %v, %v; want NONE, NONE, EOF", value, key, err)
	}
}

func TestCapabilitiesFailAfterDocumentClose(t *testing.T) {
	t.Parallel()

	ctx, root := testFSContext(t, false)
	writePDFForTest(t, root, "closed.pdf", []pdfTestPage{{Text: "Closed"}})

	document, err := Open(ctx, "closed.pdf")
	if err != nil {
		t.Fatalf("unexpected open error: %v", err)
	}
	page, err := document.Page(1)
	if err != nil {
		t.Fatalf("unexpected page error: %v", err)
	}
	pages := NewPageCollection(document)

	if err := document.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}

	if _, err := document.Get(ctx, runtime.NewString("pageCount")); err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("expected closed document property error, got %v", err)
	}
	if _, err := document.Get(ctx, runtime.NewString("pages")); err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("expected closed pages property error, got %v", err)
	}
	if _, err := page.Get(ctx, runtime.NewString("number")); err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("expected closed page property error, got %v", err)
	}
	if _, err := pages.Length(ctx); err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("expected closed page collection length error, got %v", err)
	}
	if _, err := pages.At(ctx, runtime.ZeroInt); err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("expected closed page collection index error, got %v", err)
	}
	if _, err := pages.Iterate(ctx); err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("expected closed page collection iteration error, got %v", err)
	}
}
