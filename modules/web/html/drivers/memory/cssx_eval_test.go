package memory

import (
	"context"
	"testing"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	cssxcommon "github.com/MontFerret/contrib/modules/web/html/drivers/internal/cssx"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestCompileCSSXOpsUsesCallName(t *testing.T) {
	pipeline, err := cssxcommon.Compile(`:first(div)`)
	if err != nil {
		t.Fatalf("compile cssx: %v", err)
	}

	ops, err := cssxcommon.CompilePipeline(pipeline)
	if err != nil {
		t.Fatalf("compile ops: %v", err)
	}

	if len(ops) != 2 {
		t.Fatalf("expected 2 ops, got %d", len(ops))
	}

	if ops[1].Kind != cssxcommon.OpCall {
		t.Fatalf("expected call op, got %s", ops[1].Kind)
	}

	if ops[1].Name != string(cssxcommon.ExpressionFirst) {
		t.Fatalf("expected %s, got %s", cssxcommon.ExpressionFirst, ops[1].Name)
	}
}

func TestCSSXResultNormalization(t *testing.T) {
	doc := mustDocument(t, `<div><span>one</span></div>`)
	el := &HTMLElement{doc: doc, selection: doc.Selection}

	list, err := cssxResultToList(context.Background(), el, "scalar")
	if err != nil {
		t.Fatalf("normalize scalar: %v", err)
	}

	length, err := list.Length(context.Background())
	if err != nil {
		t.Fatalf("read scalar length: %v", err)
	}

	if length != 1 {
		t.Fatalf("expected 1 scalar item")
	}

	nodes := cssxQueryAll(doc.Selection, "span")
	list, err = cssxResultToList(context.Background(), el, nodes[0])
	if err != nil {
		t.Fatalf("normalize node: %v", err)
	}

	first, err := list.At(context.Background(), runtime.NewInt(0))
	if err != nil {
		t.Fatalf("read first node item: %v", err)
	}

	typed, ok := first.(runtime.Typed)
	if !ok {
		t.Fatalf("expected typed value, got %T", first)
	}

	if typed.Type() != drivers.HTMLElementType {
		t.Fatalf("expected HTMLElement type, got %s", typed.Type())
	}
}

func TestEvalCSSXSelectionPipeline(t *testing.T) {
	ctx := context.Background()
	doc := mustDocument(t, `<main><p> A </p><p>B</p><a href="/a">A</a><a>B</a></main>`)
	el := &HTMLElement{doc: doc, selection: doc.Selection}

	texts, err := EvalCSSX(ctx, el, runtime.NewString(`:normalize(:text(p))`))
	if err != nil {
		t.Fatalf("evaluate mapped text: %v", err)
	}
	assertRuntimeList(t, texts, []runtime.Value{runtime.NewString("A"), runtime.NewString("B")})

	first, err := EvalCSSX(ctx, el, runtime.NewString(`:text(:first(p))`))
	if err != nil {
		t.Fatalf("evaluate collapsed text: %v", err)
	}
	assertRuntimeList(t, first, []runtime.Value{runtime.NewString(" A ")})

	missing, err := EvalCSSX(ctx, el, runtime.NewString(`:text(:first(.missing))`))
	if err != nil {
		t.Fatalf("evaluate missing collapsed text: %v", err)
	}
	assertRuntimeList(t, missing, nil)

	attrs, err := EvalCSSX(ctx, el, runtime.NewString(`:attr("href", a)`))
	if err != nil {
		t.Fatalf("evaluate mapped attributes: %v", err)
	}
	assertRuntimeList(t, attrs, []runtime.Value{runtime.NewString("/a"), runtime.None})

	count, err := EvalCSSX(ctx, el, runtime.NewString(`:count(:attr("href", a))`))
	if err != nil {
		t.Fatalf("count mapped attributes: %v", err)
	}
	assertRuntimeList(t, count, []runtime.Value{runtime.NewInt(2)})

	compactCount, err := EvalCSSX(ctx, el, runtime.NewString(`:count(:compact(:attr("href", a)))`))
	if err != nil {
		t.Fatalf("count compacted attributes: %v", err)
	}
	assertRuntimeList(t, compactCount, []runtime.Value{runtime.NewInt(1)})
}

func assertRuntimeList(t *testing.T, list runtime.List, expected []runtime.Value) {
	t.Helper()

	ctx := context.Background()
	length, err := list.Length(ctx)
	if err != nil {
		t.Fatalf("read list length: %v", err)
	}
	if int(length) != len(expected) {
		t.Fatalf("expected %d values, got %d", len(expected), length)
	}

	for idx, want := range expected {
		got, err := list.At(ctx, runtime.NewInt(idx))
		if err != nil {
			t.Fatalf("read item %d: %v", idx, err)
		}
		if runtime.CompareValues(got, want) != 0 {
			t.Fatalf("item %d: expected %v, got %v", idx, want, got)
		}
	}
}
