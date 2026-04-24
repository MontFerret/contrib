package dom

import (
	"context"
	"hash/fnv"

	"github.com/pkg/errors"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/contrib/modules/web/html/drivers/internal/access"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (doc *HTMLDocument) MarshalJSON() ([]byte, error) {
	return doc.element.MarshalJSON()
}

func (doc *HTMLDocument) Type() runtime.Type {
	return drivers.HTMLDocumentType
}

func (doc *HTMLDocument) String() string {
	return doc.frameTree.Frame.URL
}

func (doc *HTMLDocument) Unwrap() any {
	return doc.element
}

func (doc *HTMLDocument) Hash() uint64 {
	h := fnv.New64a()

	h.Write([]byte(doc.Type().String()))
	h.Write([]byte(":"))
	h.Write([]byte(doc.frameTree.Frame.ID))
	h.Write([]byte(doc.frameTree.Frame.URL))

	return h.Sum64()
}

func (doc *HTMLDocument) Copy() runtime.Value {
	return runtime.None
}

func (doc *HTMLDocument) Compare(other runtime.Value) int {
	switch value := other.(type) {
	case *HTMLDocument:
		thisID := runtime.NewString(string(doc.Frame().Frame.ID))
		otherID := runtime.NewString(string(value.Frame().Frame.ID))

		return thisID.Compare(otherID)
	case drivers.HTMLDocument:
		return runtime.NewString(doc.frameTree.Frame.URL).Compare(value.GetURL())
	case FrameID:
		return runtime.NewString(string(doc.frameTree.Frame.ID)).Compare(runtime.NewString(other.String()))
	default:
		return drivers.CompareTypes(doc, other)
	}
}

func (doc *HTMLDocument) Iterate(ctx context.Context) (runtime.Iterator, error) {
	return doc.element.Iterate(ctx)
}

func (doc *HTMLDocument) Get(ctx context.Context, key runtime.Value) (runtime.Value, error) {
	return access.GetInDocument(ctx, key, doc)
}

func (doc *HTMLDocument) GetNodeType(_ context.Context) (runtime.Int, error) {
	return 9, nil
}

func (doc *HTMLDocument) GetNodeName(_ context.Context) (runtime.String, error) {
	return "#document", nil
}

func (doc *HTMLDocument) GetTitle() runtime.String {
	value, err := doc.eval.EvalValue(context.Background(), templates.GetTitle())
	if err != nil {
		doc.logError(errors.Wrap(err, "failed to read document title"))

		return runtime.EmptyString
	}

	return runtime.NewString(value.String())
}

func (doc *HTMLDocument) GetName() runtime.String {
	if doc.frameTree.Frame.Name != nil {
		return runtime.NewString(*doc.frameTree.Frame.Name)
	}

	return runtime.EmptyString
}

func (doc *HTMLDocument) Length(ctx context.Context) (runtime.Int, error) {
	return doc.element.Length(ctx)
}

func (doc *HTMLDocument) GetElement() drivers.HTMLElement {
	return doc.element
}

func (doc *HTMLDocument) GetURL() runtime.String {
	return runtime.NewString(doc.frameTree.Frame.URL)
}
