package core

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestWorksheetCellsRangesAndScalarConversion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	workbook := Create()
	sheet, err := workbook.Sheet("Sheet1")
	if err != nil {
		t.Fatalf("unexpected sheet error: %v", err)
	}

	when := time.Date(2026, 7, 7, 12, 30, 0, 0, time.UTC)
	values := map[string]runtime.Value{
		"A1": runtime.NewString("Name"),
		"B1": runtime.True,
		"C1": runtime.NewInt(92),
		"D1": runtime.NewFloat(92.5),
		"E1": runtime.NewDateTime(when),
		"F1": runtime.NewString(""),
		"G1": runtime.None,
	}

	for cell, value := range values {
		if err := sheet.Set(cell, value); err != nil {
			t.Fatalf("unexpected set %s error: %v", cell, err)
		}
	}

	expected := map[string]runtime.Value{
		"A1": runtime.NewString("Name"),
		"B1": runtime.True,
		"C1": runtime.NewInt(92),
		"D1": runtime.NewFloat(92.5),
		"F1": runtime.NewString(""),
		"G1": runtime.None,
		"H1": runtime.None,
	}
	for cell, want := range expected {
		got, err := sheet.Get(cell)
		if err != nil {
			t.Fatalf("unexpected get %s error: %v", cell, err)
		}
		assertValue(t, got, want)
	}

	dateValue, err := sheet.Get("E1")
	if err != nil {
		t.Fatalf("unexpected get date error: %v", err)
	}
	if _, ok := dateValue.(runtime.DateTime); !ok {
		t.Fatalf("expected DateTime, got %T(%s)", dateValue, dateValue.String())
	}

	rows, err := sheet.Range(ctx, "A1:H1")
	if err != nil {
		t.Fatalf("unexpected range error: %v", err)
	}
	if got := listRows(t, ctx, rows); len(got) != 1 || len(got[0]) != 8 {
		t.Fatalf("unexpected range shape: %v", got)
	}
}

func TestWorksheetWriteRangeValidationAndAppend(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	workbook := Create()
	sheet, err := workbook.Sheet("Sheet1")
	if err != nil {
		t.Fatalf("unexpected sheet error: %v", err)
	}

	matrix := runtime.NewArrayWith(
		runtime.NewArrayWith(runtime.NewString("Name"), runtime.NewString("Score")),
		runtime.NewArrayWith(runtime.NewString("Alice"), runtime.NewInt(92)),
	)
	if err := sheet.WriteRange(ctx, "A1:B2", matrix); err != nil {
		t.Fatalf("unexpected write range error: %v", err)
	}

	appendRows := runtime.NewArrayWith(
		runtime.NewArrayWith(runtime.NewString("Bob"), runtime.NewInt(81)),
		runtime.NewArrayWith(runtime.NewString("Cara"), runtime.NewInt(77)),
	)
	if err := sheet.Append(ctx, appendRows); err != nil {
		t.Fatalf("unexpected append error: %v", err)
	}

	rows, err := sheet.Range(ctx, "A1:B4")
	if err != nil {
		t.Fatalf("unexpected range error: %v", err)
	}
	got := listRows(t, ctx, rows)
	assertValue(t, got[0][0], runtime.NewString("Name"))
	assertValue(t, got[1][0], runtime.NewString("Alice"))
	assertValue(t, got[2][0], runtime.NewString("Bob"))
	assertValue(t, got[3][0], runtime.NewString("Cara"))

	nonRectangular := runtime.NewArrayWith(
		runtime.NewArrayWith(runtime.NewString("Name"), runtime.NewString("Score")),
		runtime.NewArrayWith(runtime.NewString("Diana")),
	)
	err = sheet.WriteRange(ctx, "A1", nonRectangular)
	if err == nil || !strings.Contains(err.Error(), "row 2 has 1 cells; expected 2") {
		t.Fatalf("expected row width error, got %v", err)
	}

	err = sheet.WriteRange(ctx, "A1:B3", matrix)
	if err == nil || !strings.Contains(err.Error(), "expects 3 rows and 2 columns") {
		t.Fatalf("expected range shape error, got %v", err)
	}

	unsupported := runtime.NewArrayWith(runtime.NewArrayWith(runtime.NewObject()))
	err = sheet.WriteRange(ctx, "D1", unsupported)
	if err == nil || !strings.Contains(err.Error(), "unsupported XLSX cell value type") {
		t.Fatalf("expected unsupported value error, got %v", err)
	}
}

func TestWorksheetMalformedReferences(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	workbook := Create()
	sheet, err := workbook.Sheet("Sheet1")
	if err != nil {
		t.Fatalf("unexpected sheet error: %v", err)
	}

	if _, err := sheet.Get("4B"); err == nil || !strings.Contains(err.Error(), "invalid XLSX cell reference") {
		t.Fatalf("expected invalid cell error, got %v", err)
	}
	if _, err := sheet.Range(ctx, "A2:A1"); err == nil || !strings.Contains(err.Error(), "start must be above") {
		t.Fatalf("expected invalid range error, got %v", err)
	}
}
