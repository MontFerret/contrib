package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ferretfs "github.com/MontFerret/ferret/v2/pkg/fs"
)

func TestDocumentOpenAndExtractsThroughFerretFilesystem(t *testing.T) {
	t.Parallel()

	ctx, root := testFSContext(t, false)
	crop := [4]float64{0, 0, 300, 400}
	writePDFForTest(t, root, "report.pdf", []pdfTestPage{
		{Text: "Alpha", CropBox: &crop, Rotation: 90},
		{},
	})

	document, err := Open(ctx, "report.pdf")
	if err != nil {
		t.Fatalf("unexpected open error: %v", err)
	}
	t.Cleanup(func() {
		if err := document.Close(); err != nil {
			t.Fatalf("unexpected close error: %v", err)
		}
	})
	if document.source.buffer != nil {
		t.Fatal("expected root filesystem source to use random access without buffering")
	}

	count, err := document.PageCount()
	if err != nil {
		t.Fatalf("unexpected page count error: %v", err)
	}
	if count != 2 {
		t.Fatalf("page count = %d, want 2", count)
	}

	pages, err := document.Pages()
	if err != nil {
		t.Fatalf("unexpected pages error: %v", err)
	}
	if len(pages) != 2 || pages[0].Number() != 1 || pages[1].Number() != 2 {
		t.Fatalf("unexpected pages: %#v", pages)
	}

	page, err := document.Page(1)
	if err != nil {
		t.Fatalf("unexpected page error: %v", err)
	}

	text, err := page.Text(ctx)
	if err != nil {
		t.Fatalf("unexpected page text error: %v", err)
	}
	if !strings.Contains(text, "Alpha") {
		t.Fatalf("page text %q does not contain fixture text", text)
	}

	documentText, err := document.Text(ctx)
	if err != nil {
		t.Fatalf("unexpected document text error: %v", err)
	}
	if !strings.Contains(documentText, "Alpha") || !strings.Contains(documentText, "\n\n") {
		t.Fatalf("document text %q does not preserve page order with separator", documentText)
	}

	info, err := page.Info(ctx)
	if err != nil {
		t.Fatalf("unexpected page info error: %v", err)
	}
	if info.Number != 1 || info.Width != 300 || info.Height != 400 || info.Rotation != 90 {
		t.Fatalf("unexpected page info: %#v", info)
	}

	blocks, err := page.Blocks(ctx)
	if err != nil {
		t.Fatalf("unexpected blocks error: %v", err)
	}
	if !strings.Contains(blockText(blocks), "Alpha") {
		t.Fatalf("blocks did not contain fixture text: %#v", blocks)
	}
	if len(blocks) == 0 || blocks[0].Bounds.X != 72 || blocks[0].Bounds.Y != 720 || blocks[0].Bounds.Height != 24 {
		t.Fatalf("unexpected first block bounds: %#v", blocks)
	}

	emptyPage, err := document.Page(2)
	if err != nil {
		t.Fatalf("unexpected empty page error: %v", err)
	}
	emptyText, err := emptyPage.Text(ctx)
	if err != nil {
		t.Fatalf("unexpected empty page text error: %v", err)
	}
	if emptyText != "" {
		t.Fatalf("empty page text = %q, want empty", emptyText)
	}
	emptyBlocks, err := emptyPage.Blocks(ctx)
	if err != nil {
		t.Fatalf("unexpected empty blocks error: %v", err)
	}
	if len(emptyBlocks) != 0 {
		t.Fatalf("empty page blocks = %#v, want none", emptyBlocks)
	}
}

func TestPageLookupValidation(t *testing.T) {
	t.Parallel()

	ctx, root := testFSContext(t, false)
	writePDFForTest(t, root, "one.pdf", []pdfTestPage{{Text: "One"}})

	document, err := Open(ctx, "one.pdf")
	if err != nil {
		t.Fatalf("unexpected open error: %v", err)
	}
	t.Cleanup(func() { _ = document.Close() })

	if _, err := document.Page(0); err == nil || !strings.Contains(err.Error(), "out of range") {
		t.Fatalf("expected page 0 out-of-range error, got %v", err)
	}
	if _, err := document.Page(2); err == nil || !strings.Contains(err.Error(), "out of range") {
		t.Fatalf("expected page 2 out-of-range error, got %v", err)
	}
}

func TestDocumentCloseIsIdempotentAndInvalidatesPages(t *testing.T) {
	t.Parallel()

	ctx, root := testFSContext(t, false)
	writePDFForTest(t, root, "close.pdf", []pdfTestPage{{Text: "Close"}})

	document, err := Open(ctx, "close.pdf")
	if err != nil {
		t.Fatalf("unexpected open error: %v", err)
	}
	page, err := document.Page(1)
	if err != nil {
		t.Fatalf("unexpected page error: %v", err)
	}

	if document.String() != `PDFDocument("close.pdf")` {
		t.Fatalf("unexpected document display: %s", document.String())
	}
	if document.ResourceID() == 0 || document.Hash() == 0 || document.Copy() != document {
		t.Fatal("expected document to expose stable runtime resource identity")
	}
	if page.String() != "PDFPage(1)" || page.Hash() == 0 || page.Copy() != page {
		t.Fatal("expected page to expose stable runtime value identity")
	}

	if err := document.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
	if err := document.Close(); err != nil {
		t.Fatalf("unexpected idempotent close error: %v", err)
	}
	if _, err := document.PageCount(); err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("expected closed document error, got %v", err)
	}
	if _, err := page.Text(ctx); err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("expected closed page error, got %v", err)
	}
}

func TestOpenRejectsInvalidSources(t *testing.T) {
	t.Parallel()

	ctx, root := testFSContext(t, false)
	if err := os.WriteFile(filepath.Join(root, "not.pdf"), []byte("not a pdf"), 0666); err != nil {
		t.Fatalf("failed to write non-PDF: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "corrupt.pdf"), []byte("%PDF-1.4\n1 0 obj\n<<"), 0666); err != nil {
		t.Fatalf("failed to write corrupt PDF: %v", err)
	}

	for _, path := range []string{"missing.pdf", "not.pdf", "corrupt.pdf"} {
		t.Run(path, func(t *testing.T) {
			_, err := Open(ctx, path)
			if err == nil {
				t.Fatal("expected open error")
			}
		})
	}
}

func TestBufferedFallbackAndLimit(t *testing.T) {
	t.Parallel()

	data := buildPDFForTest(t, []pdfTestPage{{Text: "Buffered"}})
	ctx := ferretfs.WithFileSystem(context.Background(), memoryFS{files: map[string][]byte{"buffered.pdf": data}})

	document, err := Open(ctx, "buffered.pdf")
	if err != nil {
		t.Fatalf("unexpected buffered open error: %v", err)
	}
	if len(document.source.buffer) == 0 {
		t.Fatal("expected non-random-access source to be buffered")
	}
	t.Cleanup(func() { _ = document.Close() })

	text, err := document.Text(ctx)
	if err != nil {
		t.Fatalf("unexpected text error: %v", err)
	}
	if !strings.Contains(text, "Buffered") {
		t.Fatalf("text %q does not contain fixture text", text)
	}

	_, err = Open(ctx, "buffered.pdf", OpenOptions{MaxBufferSize: int64(len(data) - 1)})
	if err == nil || !strings.Contains(err.Error(), "buffer limit") {
		t.Fatalf("expected buffer limit error, got %v", err)
	}
}

func TestOpenReportsFilesystemFailure(t *testing.T) {
	t.Parallel()

	data := buildPDFForTest(t, []pdfTestPage{{Text: "Blocked"}})
	ctx := ferretfs.WithFileSystem(context.Background(), memoryFS{
		files:    map[string][]byte{"blocked.pdf": data},
		failOpen: true,
	})

	_, err := Open(ctx, "blocked.pdf")
	if err == nil || !strings.Contains(err.Error(), "open blocked") {
		t.Fatalf("expected filesystem open failure, got %v", err)
	}
}

func TestContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, root := testFSContext(t, false)
	writePDFForTest(t, root, "cancel.pdf", []pdfTestPage{{Text: "One"}, {Text: "Two"}})

	cancelled, cancel := context.WithCancel(ctx)
	cancel()

	if _, err := Open(cancelled, "cancel.pdf"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled open error, got %v", err)
	}

	document, err := Open(ctx, "cancel.pdf")
	if err != nil {
		t.Fatalf("unexpected open error: %v", err)
	}
	t.Cleanup(func() { _ = document.Close() })

	if _, err := document.Text(cancelled); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled text error, got %v", err)
	}
	page, err := document.Page(1)
	if err != nil {
		t.Fatalf("unexpected page error: %v", err)
	}
	if _, err := page.Blocks(cancelled); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled blocks error, got %v", err)
	}
}

func TestDocumentsRemainIsolated(t *testing.T) {
	t.Parallel()

	ctx, root := testFSContext(t, false)
	writePDFForTest(t, root, "first.pdf", []pdfTestPage{{Text: "First"}})
	writePDFForTest(t, root, "second.pdf", []pdfTestPage{{Text: "Second"}})

	first, err := Open(ctx, "first.pdf")
	if err != nil {
		t.Fatalf("unexpected first open error: %v", err)
	}
	t.Cleanup(func() { _ = first.Close() })
	second, err := Open(ctx, "second.pdf")
	if err != nil {
		t.Fatalf("unexpected second open error: %v", err)
	}
	t.Cleanup(func() { _ = second.Close() })

	firstText, err := first.Text(ctx)
	if err != nil {
		t.Fatalf("unexpected first text error: %v", err)
	}
	secondText, err := second.Text(ctx)
	if err != nil {
		t.Fatalf("unexpected second text error: %v", err)
	}

	if !strings.Contains(firstText, "First") || strings.Contains(firstText, "Second") {
		t.Fatalf("unexpected first document text: %q", firstText)
	}
	if !strings.Contains(secondText, "Second") || strings.Contains(secondText, "First") {
		t.Fatalf("unexpected second document text: %q", secondText)
	}
}
