package http

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/MontFerret/contrib/modules/web/html/drivers/common/cssx"
	cssxc "github.com/MontFerret/cssx"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

type cssxCompiledOpKind string

const (
	cssxSelectOp cssxCompiledOpKind = "select"
	cssxCallOp   cssxCompiledOpKind = "call"
)

type cssxCompiledOp struct {
	Kind     cssxCompiledOpKind
	Selector string
	Name     cssx.Expression
	Args     []any
	Arity    int
	Index    int
}

func EvalCSSX(ctx context.Context, el *HTMLElement, expression runtime.String) (runtime.List, error) {
	if el == nil || el.selection == nil {
		return runtime.NewArray(0), runtime.Error(runtime.ErrMissedArgument, "element")
	}

	pipeline, err := cssx.Compile(string(expression))

	if err != nil {
		return nil, err
	}

	ops, err := compileCSSXOps(pipeline)

	if err != nil {
		return nil, err
	}

	stack := make([]any, 0, len(ops))
	baseURL := cssxBaseURL(el.doc)

	for _, op := range ops {
		switch op.Kind {
		case cssxSelectOp:
			stack = append(stack, cssxQueryAll(el.selection, op.Selector))
		case cssxCallOp:
			consume := op.Arity

			if consume == 0 && len(stack) > 0 {
				consume = 1
			}

			if consume > len(stack) {
				stack = append(stack, []any{})
				continue
			}

			values := make([]any, consume)

			if consume > 0 {
				copy(values, stack[len(stack)-consume:])
				stack = stack[:len(stack)-consume]
			}

			stack = append(stack, cssxApplyCall(op.Name, op.Args, values, baseURL))
		default:
			return nil, fmt.Errorf("unknown operation %q", op.Kind)
		}
	}

	if len(stack) == 0 {
		return runtime.NewArray(0), nil
	}

	return cssxResultToList(ctx, el, stack[len(stack)-1])
}

func compileCSSXOps(pipeline cssxc.Pipeline) ([]cssxCompiledOp, error) {
	ops := make([]cssxCompiledOp, 0, len(pipeline.Ops))

	for idx, step := range pipeline.Ops {
		switch step.Kind {
		case cssxc.OpSelect:
			selector := strings.TrimSpace(step.Selector)

			if selector == "" {
				return nil, fmt.Errorf("invalid select operation at %d: selector is empty", idx)
			}

			ops = append(ops, cssxCompiledOp{
				Kind:     cssxSelectOp,
				Selector: selector,
				Index:    idx,
			})
		case cssxc.OpCall:
			resolved, err := cssx.ResolveSelector(step.Name)

			if err != nil {
				return nil, fmt.Errorf("invalid call operation at %d: %w", idx, err)
			}

			if err := cssx.ValidateCallArgs(resolved, step); err != nil {
				return nil, fmt.Errorf("invalid %s call at %d: %w", resolved, idx, err)
			}

			args, err := cssx.CollectCallArgs(step)

			if err != nil {
				return nil, fmt.Errorf("collect call args at %d: %w", idx, err)
			}

			ops = append(ops, cssxCompiledOp{
				Kind:  cssxCallOp,
				Name:  resolved,
				Arity: step.Arity,
				Args:  args,
				Index: idx,
			})
		default:
			return nil, fmt.Errorf("unexpected pipeline operation at %d: %d", idx, step.Kind)
		}
	}

	return ops, nil
}

func cssxResultToList(ctx context.Context, el *HTMLElement, input any) (runtime.List, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	out := runtime.NewArray(10)

	if input == nil {
		return out, nil
	}

	items, ok := input.([]any)

	if !ok {
		items = []any{input}
	}

	for _, item := range items {
		value, err := cssxToRuntimeValue(el, item)

		if err != nil {
			return nil, err
		}

		if err := out.Append(ctx, value); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func cssxToRuntimeValue(el *HTMLElement, input any) (runtime.Value, error) {
	switch v := input.(type) {
	case nil:
		return runtime.None, nil
	case *html.Node:
		doc := el.doc
		nodeDoc := goquery.NewDocumentFromNode(v)

		if doc == nil {
			doc = nodeDoc
		}

		return NewHTMLElement(doc, nodeDoc.Selection)
	case []any:
		arr := runtime.NewArray(len(v))
		ctx := context.Background()

		for _, item := range v {
			value, err := cssxToRuntimeValue(el, item)

			if err != nil {
				return runtime.None, err
			}

			if err := arr.Append(ctx, value); err != nil {
				return runtime.None, err
			}
		}

		return arr, nil
	default:
		return runtime.ValueOf(v)
	}
}

func cssxBaseURL(doc *goquery.Document) *url.URL {
	if doc == nil || doc.Url == nil {
		return nil
	}

	copy := *doc.Url

	return &copy
}
