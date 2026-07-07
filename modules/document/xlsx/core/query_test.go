package core

import (
	"context"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestWorksheetQueryRowsHeadersAndModifiers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	workbook := Create()
	sheet, err := workbook.Sheet("Sheet1")
	if err != nil {
		t.Fatalf("unexpected sheet error: %v", err)
	}

	rows := runtime.NewArrayWith(
		runtime.NewArrayWith(runtime.NewString("Name"), runtime.NewString("Name"), runtime.None, runtime.NewString("Active")),
		runtime.NewArrayWith(runtime.NewString("Alice"), runtime.NewString("Engineering"), runtime.NewString("A"), runtime.True),
		runtime.NewArrayWith(runtime.NewString("Bob"), runtime.NewString("Finance"), runtime.NewString("B"), runtime.False),
		runtime.NewArrayWith(runtime.None, runtime.None, runtime.None, runtime.None),
	)
	if err := sheet.WriteRange(ctx, "A1:D4", rows); err != nil {
		t.Fatalf("unexpected write range error: %v", err)
	}

	out, err := sheet.Query(ctx, runtime.Query{Expression: runtime.NewString("A1:D4")})
	if err != nil {
		t.Fatalf("unexpected query error: %v", err)
	}
	if got := listRows(t, ctx, out); len(got) != 4 {
		t.Fatalf("expected 4 row arrays, got %d", len(got))
	}

	withHeaders := runtime.NewObjectWith(map[string]runtime.Value{
		"headers":       runtime.True,
		"trimEmptyRows": runtime.True,
	})
	out, err = sheet.Query(ctx, runtime.Query{
		Expression: runtime.NewString("A1:D4"),
		Params:     withHeaders,
	})
	if err != nil {
		t.Fatalf("unexpected query headers error: %v", err)
	}

	length, err := out.Length(ctx)
	if err != nil {
		t.Fatalf("unexpected length error: %v", err)
	}
	if length != 2 {
		t.Fatalf("expected 2 data rows after trim, got %d", length)
	}

	first, err := out.At(ctx, runtime.ZeroInt)
	if err != nil {
		t.Fatalf("unexpected first row error: %v", err)
	}
	assertValue(t, objectField(t, ctx, first, "Name"), runtime.NewString("Alice"))
	assertValue(t, objectField(t, ctx, first, "Name_2"), runtime.NewString("Engineering"))
	assertValue(t, objectField(t, ctx, first, "column_3"), runtime.NewString("A"))
	assertValue(t, objectField(t, ctx, first, "Active"), runtime.True)

	one, err := sheet.QueryOne(ctx, runtime.Query{Expression: runtime.NewString("A2:D3")})
	if err != nil {
		t.Fatalf("unexpected query one error: %v", err)
	}
	row, ok := one.(runtime.List)
	if !ok {
		t.Fatalf("expected row array, got %T", one)
	}
	name, err := row.At(ctx, runtime.ZeroInt)
	if err != nil {
		t.Fatalf("unexpected first cell error: %v", err)
	}
	assertValue(t, name, runtime.NewString("Alice"))

	count, err := sheet.QueryCount(ctx, runtime.Query{Expression: runtime.NewString("A2:D3")})
	if err != nil {
		t.Fatalf("unexpected query count error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected count 2, got %d", count)
	}

	exists, err := sheet.QueryExists(ctx, runtime.Query{Expression: runtime.NewString("A2:D3")})
	if err != nil {
		t.Fatalf("unexpected query exists error: %v", err)
	}
	if !exists {
		t.Fatal("expected query exists")
	}
}

func TestWorksheetQueryValidation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	workbook := Create()
	sheet, err := workbook.Sheet("Sheet1")
	if err != nil {
		t.Fatalf("unexpected sheet error: %v", err)
	}

	_, err = sheet.Query(ctx, runtime.Query{
		Expression: runtime.NewString("A1:A1"),
		Kind:       runtime.NewString("xlsx"),
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported XLSX query dialect") {
		t.Fatalf("expected unsupported dialect error, got %v", err)
	}

	_, err = sheet.Query(ctx, runtime.Query{
		Expression: runtime.NewString("A1:A1"),
		Params:     runtime.NewObjectWith(map[string]runtime.Value{"headers": runtime.NewString("yes")}),
	})
	if err == nil || !strings.Contains(err.Error(), "WITH.headers must be a boolean") {
		t.Fatalf("expected invalid WITH error, got %v", err)
	}

	_, err = sheet.Query(ctx, runtime.Query{
		Expression: runtime.NewString("A1:A1"),
		Params:     runtime.NewObjectWith(map[string]runtime.Value{"limit": runtime.NewInt(10)}),
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported XLSX query WITH.limit") {
		t.Fatalf("expected unsupported WITH key error, got %v", err)
	}
}
