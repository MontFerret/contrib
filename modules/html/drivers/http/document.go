package http

import (
	"context"
	"hash/fnv"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/contrib/modules/html/drivers/common"
	"github.com/PuerkitoBio/goquery"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type HTMLDocument struct {
	doc      *goquery.Document
	element  drivers.HTMLElement
	url      runtime.String
	parent   drivers.HTMLDocument
	children runtime.List
}

func NewRootHTMLDocument(
	node *goquery.Document,
	url string,
) (*HTMLDocument, error) {
	return NewHTMLDocument(node, url, nil)
}

func NewHTMLDocument(
	node *goquery.Document,
	url string,
	parent drivers.HTMLDocument,
) (*HTMLDocument, error) {
	if url == "" {
		return nil, runtime.Error(runtime.ErrMissedArgument, "document url")
	}

	if node == nil {
		return nil, runtime.Error(runtime.ErrMissedArgument, "document root selection")
	}

	el, err := NewHTMLElement(node.Selection)

	if err != nil {
		return nil, err
	}

	doc := new(HTMLDocument)
	doc.doc = node
	doc.element = el
	doc.parent = parent
	doc.url = runtime.NewString(url)
	doc.children = runtime.NewArray(10)

	frames := node.Find("iframe")
	frames.Each(func(_ int, selection *goquery.Selection) {
		child, _ := NewHTMLDocument(goquery.NewDocumentFromNode(selection.Nodes[0]), selection.AttrOr("src", url), doc)

		_ = doc.children.Append(context.Background(), child)
	})

	return doc, nil
}

func (doc *HTMLDocument) MarshalJSON() ([]byte, error) {
	return doc.element.MarshalJSON()
}

func (doc *HTMLDocument) Type() runtime.Type {
	return drivers.HTMLDocumentType
}

func (doc *HTMLDocument) String() string {
	str, err := doc.doc.Html()

	if err != nil {
		return ""
	}

	return str
}

func (doc *HTMLDocument) Compare(other runtime.Value) int64 {
	switch other.Type() {
	case drivers.HTMLElementType:
		otherDoc := other.(drivers.HTMLDocument)

		return doc.url.Compare(otherDoc.GetURL())
	default:
		return drivers.Compare(doc.Type(), other.Type())
	}
}

func (doc *HTMLDocument) Unwrap() any {
	return doc.doc
}

func (doc *HTMLDocument) Hash() uint64 {
	h := fnv.New64a()

	h.Write([]byte(doc.Type().String()))
	h.Write([]byte(":"))
	h.Write([]byte(doc.url))

	return h.Sum64()
}

func (doc *HTMLDocument) Copy() runtime.Value {
	cp, err := NewHTMLDocument(doc.doc, string(doc.url), doc.parent)

	if err != nil {
		return runtime.None
	}

	return cp
}

func (doc *HTMLDocument) Clone(_ context.Context) (runtime.Cloneable, error) {
	cloned, err := NewHTMLDocument(doc.doc, doc.url.String(), doc.parent)

	if err != nil {
		return runtime.None, err
	}

	return cloned, nil
}

func (doc *HTMLDocument) Length(_ context.Context) (runtime.Int, error) {
	return runtime.NewInt(doc.doc.Length()), nil
}

func (doc *HTMLDocument) Iterate(_ context.Context) (runtime.Iterator, error) {
	return common.NewIterator(doc.element)
}

func (doc *HTMLDocument) Get(ctx context.Context, key runtime.Value) (runtime.Value, error) {
	return common.GetInDocument(ctx, key, doc)
}

func (doc *HTMLDocument) Set(ctx context.Context, key runtime.Value, value runtime.Value) error {
	return common.SetInDocument(ctx, key, doc, value)
}

func (doc *HTMLDocument) GetNodeType(_ context.Context) (runtime.Int, error) {
	return 9, nil
}

func (doc *HTMLDocument) GetNodeName(_ context.Context) (runtime.String, error) {
	return "#document", nil
}

func (doc *HTMLDocument) GetChildNodes(ctx context.Context) (runtime.List, error) {
	return doc.element.GetChildNodes(ctx)
}

func (doc *HTMLDocument) GetChildNode(ctx context.Context, idx runtime.Int) (runtime.Value, error) {
	return doc.element.GetChildNode(ctx, idx)
}

func (doc *HTMLDocument) QuerySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.Value, error) {
	return doc.element.QuerySelector(ctx, selector)
}

func (doc *HTMLDocument) QuerySelectorAll(ctx context.Context, selector drivers.QuerySelector) (runtime.List, error) {
	return doc.element.QuerySelectorAll(ctx, selector)
}

func (doc *HTMLDocument) CountBySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.Int, error) {
	return doc.element.CountBySelector(ctx, selector)
}

func (doc *HTMLDocument) ExistsBySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.Boolean, error) {
	return doc.element.ExistsBySelector(ctx, selector)
}

func (doc *HTMLDocument) XPath(ctx context.Context, expression runtime.String) (runtime.Value, error) {
	return doc.element.XPath(ctx, expression)
}

func (doc *HTMLDocument) GetTitle() runtime.String {
	title := doc.doc.Find("head > title")

	return runtime.NewString(title.Text())
}

func (doc *HTMLDocument) GetChildDocuments(ctx context.Context) (runtime.List, error) {
	cloned, err := doc.children.Clone(ctx)

	if err != nil {
		return nil, err
	}

	return runtime.CastList(cloned)
}

func (doc *HTMLDocument) GetURL() runtime.String {
	return doc.url
}

func (doc *HTMLDocument) GetElement() drivers.HTMLElement {
	return doc.element
}

func (doc *HTMLDocument) GetName() runtime.String {
	return ""
}

func (doc *HTMLDocument) GetParentDocument(_ context.Context) (drivers.HTMLDocument, error) {
	return doc.parent, nil
}

func (doc *HTMLDocument) ScrollTop(_ context.Context, _ drivers.ScrollOptions) error {
	return runtime.ErrNotSupported
}

func (doc *HTMLDocument) ScrollBottom(_ context.Context, _ drivers.ScrollOptions) error {
	return runtime.ErrNotSupported
}

func (doc *HTMLDocument) ScrollBySelector(_ context.Context, _ drivers.QuerySelector, _ drivers.ScrollOptions) error {
	return runtime.ErrNotSupported
}

func (doc *HTMLDocument) Scroll(_ context.Context, _ drivers.ScrollOptions) error {
	return runtime.ErrNotSupported
}

func (doc *HTMLDocument) MoveMouseByXY(_ context.Context, _, _ runtime.Float) error {
	return runtime.ErrNotSupported
}

func (doc *HTMLDocument) Close() error {
	return nil
}

func (doc *HTMLDocument) Query(ctx context.Context, q runtime.Query) (runtime.Value, error) {
	//TODO implement me
	panic("implement me")
}

func (doc *HTMLDocument) Dispatch(ctx context.Context, event runtime.DispatchEvent) (runtime.Value, error) {
	//TODO implement me
	panic("implement me")
}
