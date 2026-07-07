package pdf

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/source"
)

func TestNewSmoke(t *testing.T) {
	mod := New()

	if mod == nil {
		t.Fatal("expected module to be non-nil")
	}
	if mod.Name() != "document/pdf" {
		t.Fatalf("expected module name %q, got %q", "document/pdf", mod.Name())
	}
}

func TestModuleRunsPDFWorkflowFromFQL(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := "report.pdf"
	writePDFForTest(t, root, path, []pdfTestPage{
		{Text: "Alpha", Width: 612, Height: 792},
		{Text: "Beta", Width: 400, Height: 500, X: 36, Y: 420, FontSize: 18},
	})

	engine, err := ferret.New(
		ferret.WithModules(New()),
		ferret.WithFSRoot(root),
		ferret.WithRuntimeParam("path", runtime.NewString(path)),
	)
	if err != nil {
		t.Fatalf("unexpected engine error: %v", err)
	}
	t.Cleanup(func() {
		if err := engine.Close(); err != nil {
			t.Fatalf("unexpected engine close error: %v", err)
		}
	})

	output, err := engine.Run(context.Background(), source.NewAnonymous(`
		LET document = DOCUMENT::PDF::OPEN(@path)
		LET pages = DOCUMENT::PDF::PAGES(document)
		LET first = DOCUMENT::PDF::PAGE(document, 1)
		LET second = pages[1]
		LET result = {
			count: DOCUMENT::PDF::PAGE_COUNT(document),
			firstText: DOCUMENT::PDF::TEXT(first),
			fullText: DOCUMENT::PDF::TEXT(document),
			firstInfo: DOCUMENT::PDF::PAGE_INFO(first),
			secondInfo: DOCUMENT::PDF::PAGE_INFO(second),
			blocks: DOCUMENT::PDF::BLOCKS(first),
			closed: DOCUMENT::PDF::CLOSE(document)
		}
		RETURN result
	`))
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}

	var actual struct {
		FirstText string `json:"firstText"`
		FullText  string `json:"fullText"`
		Blocks    []struct {
			Text   string `json:"text"`
			Bounds struct {
				X      float64 `json:"x"`
				Y      float64 `json:"y"`
				Height float64 `json:"height"`
			} `json:"bounds"`
		} `json:"blocks"`
		FirstInfo  info `json:"firstInfo"`
		SecondInfo info `json:"secondInfo"`
		Count      int  `json:"count"`
		Closed     bool `json:"closed"`
	}
	if err := json.Unmarshal(output.Content, &actual); err != nil {
		t.Fatalf("failed to decode output: %v", err)
	}

	if actual.Count != 2 || !actual.Closed {
		t.Fatalf("unexpected count/closed result: %#v", actual)
	}
	if !strings.Contains(actual.FirstText, "Alpha") || !strings.Contains(actual.FullText, "Beta") || !strings.Contains(actual.FullText, "\n\n") {
		t.Fatalf("unexpected extracted text: %#v", actual)
	}
	if actual.FirstInfo.Number != 1 || actual.FirstInfo.Width != 612 || actual.FirstInfo.Height != 792 {
		t.Fatalf("unexpected first page info: %#v", actual.FirstInfo)
	}
	if actual.SecondInfo.Number != 2 || actual.SecondInfo.Width != 400 || actual.SecondInfo.Height != 500 {
		t.Fatalf("unexpected second page info: %#v", actual.SecondInfo)
	}
	if len(actual.Blocks) == 0 || actual.Blocks[0].Bounds.X != 72 || actual.Blocks[0].Bounds.Y != 720 || actual.Blocks[0].Bounds.Height != 24 {
		t.Fatalf("unexpected positioned blocks: %#v", actual.Blocks)
	}
}

func TestModuleReportsWrongArgumentTypesFromFQL(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := "report.pdf"
	writePDFForTest(t, root, path, []pdfTestPage{{Text: "Alpha"}})

	engine, err := ferret.New(
		ferret.WithModules(New()),
		ferret.WithFSRoot(root),
		ferret.WithRuntimeParam("path", runtime.NewString(path)),
	)
	if err != nil {
		t.Fatalf("unexpected engine error: %v", err)
	}
	t.Cleanup(func() {
		if err := engine.Close(); err != nil {
			t.Fatalf("unexpected engine close error: %v", err)
		}
	})

	output, err := engine.Run(context.Background(), source.NewAnonymous(`
		LET document = DOCUMENT::PDF::OPEN(@path)
		LET page = DOCUMENT::PDF::PAGE(document, 1)
		LET openErr = DOCUMENT::PDF::OPEN(1) ON ERROR RETURN "open"
		LET pageCountErr = DOCUMENT::PDF::PAGE_COUNT("x") ON ERROR RETURN "page_count"
		LET pagesErr = DOCUMENT::PDF::PAGES("x") ON ERROR RETURN "pages"
		LET pageDocumentErr = DOCUMENT::PDF::PAGE("x", 1) ON ERROR RETURN "page_document"
		LET pageNumberErr = DOCUMENT::PDF::PAGE(document, "1") ON ERROR RETURN "page_number"
		LET textErr = DOCUMENT::PDF::TEXT("x") ON ERROR RETURN "text"
		LET pageInfoErr = DOCUMENT::PDF::PAGE_INFO(document) ON ERROR RETURN "page_info"
		LET blocksErr = DOCUMENT::PDF::BLOCKS(document) ON ERROR RETURN "blocks"
		LET closeErr = DOCUMENT::PDF::CLOSE(page) ON ERROR RETURN "close"
		DOCUMENT::PDF::CLOSE(document)

		RETURN openErr == "open"
			AND pageCountErr == "page_count"
			AND pagesErr == "pages"
			AND pageDocumentErr == "page_document"
			AND pageNumberErr == "page_number"
			AND textErr == "text"
			AND pageInfoErr == "page_info"
			AND blocksErr == "blocks"
			AND closeErr == "close"
	`))
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}

	var actual bool
	if err := json.Unmarshal(output.Content, &actual); err != nil {
		t.Fatalf("failed to decode output: %v", err)
	}
	if !actual {
		t.Fatal("expected wrong argument workflow to return true")
	}
}

type info struct {
	Number   int     `json:"number"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	Rotation int     `json:"rotation"`
}
