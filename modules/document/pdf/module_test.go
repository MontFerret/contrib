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
		LET pages = document.pages
		LET first = pages[0]
		LET second = pages[1]
		LET result = {
			count: document.pageCount,
			length: LENGTH(pages),
			firstText: first.text,
			combinedText: (
				FOR page IN pages
					RETURN page.text
			),
			firstInfo: {
				number: first.number,
				width: first.width,
				height: first.height,
				rotation: first.rotation
			},
			secondInfo: {
				number: second.number,
				width: second.width,
				height: second.height,
				rotation: second.rotation
			},
			blocks: first.blocks,
			unknownDocument: document.missing,
			unknownPage: first.missing
		}
		RETURN result
	`))
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}

	var actual struct {
		UnknownDocument any      `json:"unknownDocument"`
		UnknownPage     any      `json:"unknownPage"`
		FirstText       string   `json:"firstText"`
		CombinedText    []string `json:"combinedText"`
		Blocks          []struct {
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
		Length     int  `json:"length"`
	}
	if err := json.Unmarshal(output.Content, &actual); err != nil {
		t.Fatalf("failed to decode output: %v", err)
	}

	if actual.Count != 2 || actual.Length != 2 {
		t.Fatalf("unexpected count/length result: %#v", actual)
	}
	if !strings.Contains(actual.FirstText, "Alpha") || len(actual.CombinedText) != 2 || !strings.Contains(actual.CombinedText[1], "Beta") {
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
	if actual.UnknownDocument != nil || actual.UnknownPage != nil {
		t.Fatalf("expected unknown properties to encode as null, got %#v", actual)
	}
}

func TestModuleReportsCapabilityErrorsFromFQL(t *testing.T) {
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
		LET openErr = DOCUMENT::PDF::OPEN(1) ON ERROR RETURN "open"
		LET missingDocument = document.missing
		LET missingPage = document.pages[0].missing
		LET outOfRange = document.pages[99]

		RETURN openErr == "open"
			AND missingDocument == NONE
			AND missingPage == NONE
			AND outOfRange == NONE
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
