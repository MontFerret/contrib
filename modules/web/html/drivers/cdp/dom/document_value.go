package dom

import (
	"context"
	"hash/fnv"

	"github.com/pkg/errors"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/contrib/modules/web/html/drivers/internal/data"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (doc *HTMLDocument) MarshalJSON() ([]byte, error) {
	return withDocumentResult(context.Background(), doc, func(state *documentState) ([]byte, error) {
		return state.element.MarshalJSON()
	})
}

func (doc *HTMLDocument) Type() runtime.Type {
	return drivers.HTMLDocumentType
}

func (doc *HTMLDocument) String() string {
	return doc.Frame().Frame.URL
}

func (doc *HTMLDocument) Unwrap() any {
	return doc.currentState().element
}

func (doc *HTMLDocument) Hash() uint64 {
	h := fnv.New64a()

	h.Write([]byte(doc.Type().String()))
	h.Write([]byte(":"))
	h.Write([]byte(doc.identity))

	return h.Sum64()
}

func (doc *HTMLDocument) Copy() runtime.Value {
	return runtime.None
}

func (doc *HTMLDocument) Equal(ctx context.Context, other runtime.Value) (bool, error) {
	if _, err := checkComparisonContext(ctx); err != nil {
		return false, err
	}

	otherDocument, ok := other.(*HTMLDocument)
	if !ok {
		return false, nil
	}

	comparison, err := doc.compare(ctx, otherDocument)

	return comparison == runtime.Equal, err
}

func (doc *HTMLDocument) Compare(ctx context.Context, other runtime.Value) (runtime.Ordering, error) {
	if comparison, err := checkComparisonContext(ctx); err != nil {
		return comparison, err
	}

	otherDocument, ok := other.(*HTMLDocument)
	if ok {
		return doc.compare(ctx, otherDocument)
	}

	if _, ok := other.(drivers.HTMLDocument); ok {
		if comparison, err := checkComparisonContext(ctx); err != nil {
			return comparison, err
		}

		return runtime.Greater, nil
	}

	return invalidComparison(doc, other)
}

func (doc *HTMLDocument) compare(ctx context.Context, other *HTMLDocument) (runtime.Ordering, error) {
	if comparison, err := checkComparisonContext(ctx); err != nil {
		return comparison, err
	}

	return runtime.CompareValues(
		ctx,
		runtime.NewString(string(doc.identity)),
		runtime.NewString(string(other.identity)),
	)
}

func (doc *HTMLDocument) Iterate(ctx context.Context) (runtime.Iterator, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.Iterator, error) {
		return state.element.Iterate(ctx)
	})
}

func (doc *HTMLDocument) Get(ctx context.Context, key runtime.Value) (runtime.Value, error) {
	return data.GetInDocument(ctx, key, doc)
}

func (doc *HTMLDocument) GetNodeType(_ context.Context) (runtime.Int, error) {
	return 9, nil
}

func (doc *HTMLDocument) GetNodeName(_ context.Context) (runtime.String, error) {
	return "#document", nil
}

func (doc *HTMLDocument) GetTitle() runtime.String {
	value, err := withDocumentResult(context.Background(), doc, func(state *documentState) (runtime.Value, error) {
		return state.eval.EvalValue(context.Background(), templates.GetTitle())
	})
	if err != nil {
		doc.logError(errors.Wrap(err, "failed to read document title"))

		return runtime.EmptyString
	}

	return runtime.NewString(value.String())
}

func (doc *HTMLDocument) GetName() runtime.String {
	frame := doc.Frame().Frame
	if frame.Name != nil {
		return runtime.NewString(*frame.Name)
	}

	return runtime.EmptyString
}

func (doc *HTMLDocument) Length(ctx context.Context) (runtime.Int, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.Int, error) {
		return state.element.Length(ctx)
	})
}

func (doc *HTMLDocument) GetElement() drivers.HTMLElement {
	return doc.currentState().element
}

func (doc *HTMLDocument) GetURL() runtime.String {
	return runtime.NewString(doc.Frame().Frame.URL)
}

func (doc *HTMLDocument) GetCurrentURL(ctx context.Context) (runtime.String, error) {
	value, err := withDocumentResult(ctx, doc, func(state *documentState) (runtime.Value, error) {
		return state.eval.EvalValue(ctx, templates.GetDocumentURL())
	})
	if err != nil {
		return runtime.EmptyString, err
	}

	return runtime.ToString(value), nil
}

func (doc *HTMLDocument) GetBaseURL(ctx context.Context) (runtime.String, error) {
	value, err := withDocumentResult(ctx, doc, func(state *documentState) (runtime.Value, error) {
		return state.eval.EvalValue(ctx, templates.GetBaseURL())
	})
	if err != nil {
		return runtime.EmptyString, err
	}

	return runtime.ToString(value), nil
}

func (doc *HTMLDocument) ResolveURL(ctx context.Context, url runtime.String) (runtime.String, error) {
	value, err := withDocumentResult(ctx, doc, func(state *documentState) (runtime.Value, error) {
		return state.eval.EvalValue(ctx, templates.ResolveURL(url))
	})
	if err != nil {
		return runtime.EmptyString, err
	}

	return runtime.ToString(value), nil
}
