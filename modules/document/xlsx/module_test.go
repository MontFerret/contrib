package xlsx

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk/sdktest"
)

func TestNewSmoke(t *testing.T) {
	mod := New()

	if mod == nil {
		t.Fatal("expected module to be non-nil")
	}
	if mod.Name() != "document/xlsx" {
		t.Fatalf("expected module name %q, got %q", "document/xlsx", mod.Name())
	}
}

func TestModuleRunsWorkbookWorkflowFromFQL(t *testing.T) {
	t.Parallel()

	harness := sdktest.New(t, ferret.WithModules(New()))

	output, err := harness.Run(context.Background(), `
		LET workbook = DOCUMENT::XLSX::CREATE()
		LET source = DOCUMENT::XLSX::SHEET(workbook, "Sheet1")

		DOCUMENT::XLSX::WRITE_RANGE(source, "A1:C4", [
			["Name", "Active", "Score"],
			["Alice", true, 92],
			["Bob", false, 81],
			[NONE, NONE, NONE]
		])

		LET rows = QUERY "A1:C4" IN source WITH {
			headers: true,
			trimEmptyRows: true
		}
		LET first = QUERY ONE "A2:C3" IN source
		LET count = QUERY COUNT "A2:C3" IN source
		LET exists = QUERY EXISTS "A2:C3" IN source

		RETURN rows[0].Name == "Alice"
			AND rows[0].Active == true
			AND rows[0].Score == 92
			AND source[~ "A2:C3"][1][0] == "Bob"
			AND first[0] == "Alice"
			AND count == 2
			AND exists
			AND DOCUMENT::XLSX::CLOSE(workbook)
	`)
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}

	var actual bool
	if err := json.Unmarshal(output.Content, &actual); err != nil {
		t.Fatalf("failed to decode output: %v", err)
	}
	if !actual {
		t.Fatal("expected FQL workflow to return true")
	}
}

func TestModuleSavesAndReopensWorkbookFromFQL(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := "report.xlsx"
	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithFSRoot(root),
		ferret.WithRuntimeParam("path", runtime.NewString(path)),
	)
	output, err := harness.Run(context.Background(), `
		LET workbook = DOCUMENT::XLSX::CREATE()
		LET sheet = DOCUMENT::XLSX::SHEET(workbook, "Sheet1")
		DOCUMENT::XLSX::SET(sheet, "A1", "Saved")
		DOCUMENT::XLSX::SAVE_AS(workbook, @path)
		DOCUMENT::XLSX::SET(sheet, "A2", "Again")
		DOCUMENT::XLSX::SAVE(workbook)
		DOCUMENT::XLSX::CLOSE(workbook)

		LET reopened = DOCUMENT::XLSX::OPEN(@path)
		LET reopenedSheet = DOCUMENT::XLSX::SHEET(reopened, "Sheet1")
		LET ok = DOCUMENT::XLSX::GET(reopenedSheet, "A1") == "Saved"
			AND DOCUMENT::XLSX::GET(reopenedSheet, "A2") == "Again"
		DOCUMENT::XLSX::CLOSE(reopened)
		RETURN ok
	`)
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}

	var actual bool
	if err := json.Unmarshal(output.Content, &actual); err != nil {
		t.Fatalf("failed to decode output: %v", err)
	}
	if !actual {
		t.Fatal("expected FQL save/reopen workflow to return true")
	}
}

func TestModuleSaveAsHonorsReadOnlyFSFromFQL(t *testing.T) {
	t.Parallel()

	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithFSRoot(t.TempDir()),
		ferret.WithFSReadOnly(),
	)

	_, err := harness.Run(context.Background(), `
		LET workbook = DOCUMENT::XLSX::CREATE()
		DOCUMENT::XLSX::SAVE_AS(workbook, "blocked.xlsx")
		RETURN true
	`)
	if err == nil || !strings.Contains(err.Error(), "read-only") {
		t.Fatalf("expected read-only filesystem error, got %v", err)
	}
}
