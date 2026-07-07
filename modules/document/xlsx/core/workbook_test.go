package core

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestWorkbookLifecycleSheetsAndStaleHandles(t *testing.T) {
	t.Parallel()

	workbook := Create()
	if workbook.String() != `XLSXWorkbook("<memory>")` {
		t.Fatalf("unexpected display: %s", workbook.String())
	}

	sheets, err := workbook.Sheets()
	if err != nil {
		t.Fatalf("unexpected sheets error: %v", err)
	}
	if len(sheets) != 1 || sheets[0] != "Sheet1" {
		t.Fatalf("unexpected sheets: %v", sheets)
	}

	sheet, err := workbook.AddSheet("Sales")
	if err != nil {
		t.Fatalf("unexpected add sheet error: %v", err)
	}
	if sheet.String() != `XLSXSheet("Sales")` {
		t.Fatalf("unexpected sheet display: %s", sheet.String())
	}
	if _, err := workbook.AddSheet("Sales"); err == nil {
		t.Fatal("expected duplicate sheet error")
	}
	if err := workbook.DeleteSheet("Sales"); err != nil {
		t.Fatalf("unexpected delete sheet error: %v", err)
	}

	if _, err := sheet.Get("A1"); err == nil || !strings.Contains(err.Error(), "deleted") {
		t.Fatalf("expected deleted sheet error, got %v", err)
	}

	recreated, err := workbook.AddSheet("Sales")
	if err != nil {
		t.Fatalf("unexpected recreate sheet error: %v", err)
	}
	if err := recreated.Set("A1", runtime.NewString("ok")); err != nil {
		t.Fatalf("unexpected set error: %v", err)
	}
	if _, err := sheet.Get("A1"); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("expected stale sheet error, got %v", err)
	}

	if err := workbook.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
	if err := workbook.Close(); err != nil {
		t.Fatalf("unexpected idempotent close error: %v", err)
	}
	if _, err := workbook.Sheets(); err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("expected closed workbook error, got %v", err)
	}
	if _, err := recreated.Get("A1"); err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("expected closed sheet error, got %v", err)
	}
}

func TestWorkbookSaveAsOpenAndSave(t *testing.T) {
	t.Parallel()

	ctx, _ := testFSContext(t, false)
	workbook := Create()
	sheet, err := workbook.Sheet("Sheet1")
	if err != nil {
		t.Fatalf("unexpected sheet error: %v", err)
	}
	if err := sheet.Set("A1", runtime.NewString("Name")); err != nil {
		t.Fatalf("unexpected set A1 error: %v", err)
	}
	if err := sheet.Set("B1", runtime.NewInt(42)); err != nil {
		t.Fatalf("unexpected set B1 error: %v", err)
	}

	if err := workbook.Save(ctx); err == nil || !strings.Contains(err.Error(), "SAVE_AS") {
		t.Fatalf("expected save without path error, got %v", err)
	}
	if err := workbook.SaveAs(ctx, "missing/out.xlsx"); err == nil || !strings.Contains(err.Error(), "parent directory") {
		t.Fatalf("expected missing parent error, got %v", err)
	}

	path := "report.xlsx"
	if err := workbook.SaveAs(ctx, path); err != nil {
		t.Fatalf("unexpected save as error: %v", err)
	}
	if err := workbook.Save(ctx); err != nil {
		t.Fatalf("unexpected save error: %v", err)
	}
	if err := workbook.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}

	reopened, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("unexpected open error: %v", err)
	}
	t.Cleanup(func() {
		if err := reopened.Close(); err != nil {
			t.Fatalf("unexpected close error: %v", err)
		}
	})

	reopenedSheet, err := reopened.Sheet("Sheet1")
	if err != nil {
		t.Fatalf("unexpected reopened sheet error: %v", err)
	}

	value, err := reopenedSheet.Get("A1")
	if err != nil {
		t.Fatalf("unexpected get A1 error: %v", err)
	}
	assertValue(t, value, runtime.NewString("Name"))

	value, err = reopenedSheet.Get("B1")
	if err != nil {
		t.Fatalf("unexpected get B1 error: %v", err)
	}
	assertValue(t, value, runtime.NewInt(42))
}

func TestOpenMissingWorkbook(t *testing.T) {
	t.Parallel()

	ctx, _ := testFSContext(t, false)
	_, err := Open(ctx, "missing.xlsx")
	if err == nil || !strings.Contains(err.Error(), "file does not exist") {
		t.Fatalf("expected missing workbook error, got %v", err)
	}
}

func TestWorkbookUsesFerretFilesystemRoot(t *testing.T) {
	t.Parallel()

	ctx, root := testFSContext(t, false)
	outsideName := "outside-" + filepath.Base(root) + ".xlsx"
	outsidePath := filepath.Join(filepath.Dir(root), outsideName)

	workbook := Create()
	err := workbook.SaveAs(ctx, "../"+outsideName)
	if err == nil {
		t.Fatal("expected save outside filesystem root to fail")
	}
	if _, statErr := os.Stat(outsidePath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected outside file to be absent, stat error: %v", statErr)
	}
}

func TestWorkbookHonorsReadOnlyFerretFilesystem(t *testing.T) {
	t.Parallel()

	ctx, _ := testFSContext(t, true)
	workbook := Create()

	err := workbook.SaveAs(ctx, "report.xlsx")
	if err == nil || !strings.Contains(err.Error(), "read-only") {
		t.Fatalf("expected read-only filesystem error, got %v", err)
	}
}
