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
