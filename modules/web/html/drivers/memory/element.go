package memory

import (
	"context"
	"hash/fnv"
	"strings"

	"golang.org/x/net/html"

	"github.com/PuerkitoBio/goquery"
	"github.com/goccy/go-json"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/internal/access"
	"github.com/MontFerret/contrib/modules/web/html/drivers/internal/nodeutil"
	"github.com/MontFerret/contrib/modules/web/html/drivers/internal/queryutil"
	"github.com/MontFerret/contrib/modules/web/html/internal/styleutil"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type HTMLElement struct {
	doc       *goquery.Document
	selection *goquery.Selection
	attrs     *runtime.Object
	styles    *runtime.Object
	children  *runtime.Array
}

func NewHTMLElement(doc *goquery.Document, node *goquery.Selection) (drivers.HTMLElement, error) {
	if node == nil {
		return nil, runtime.Error(runtime.ErrMissedArgument, "element selection")
	}

	return &HTMLElement{doc, node, nil, nil, nil}, nil
}

func (el *HTMLElement) MarshalJSON() ([]byte, error) {
	return json.Marshal(el.String())
}

func (el *HTMLElement) Type() runtime.Type {
	return drivers.HTMLElementType
}

func (el *HTMLElement) String() string {
	ih, err := el.GetInnerHTML(context.Background())

	if err != nil {
		return ""
	}

	return ih.String()
}

func (el *HTMLElement) Compare(other runtime.Value) int {
	otherElement, ok := other.(drivers.HTMLElement)

	if !ok {
		typed, ok := other.(runtime.Typed)

		if !ok {
			return 1
		}

		return drivers.Compare(el.Type(), typed.Type())
	}

	return strings.Compare(el.String(), otherElement.String())
}

func (el *HTMLElement) Unwrap() any {
	return el.selection
}

func (el *HTMLElement) Hash() uint64 {
	str, err := el.selection.Html()

	if err != nil {
		return 0
	}

	h := fnv.New64a()

	h.Write([]byte(el.Type().String()))
	h.Write([]byte(":"))
	h.Write([]byte(str))

	return h.Sum64()
}

func (el *HTMLElement) Copy() runtime.Value {
	c, _ := NewHTMLElement(el.doc, el.selection.Clone())

	return c
}

func (el *HTMLElement) GetNodeType(_ context.Context) (runtime.Int, error) {
	nodes := el.selection.Nodes

	if len(nodes) == 0 {
		return 0, nil
	}

	return runtime.NewInt(nodeutil.FromHTMLType(nodes[0].Type)), nil
}

func (el *HTMLElement) Close() error {
	return nil
}

func (el *HTMLElement) GetNodeName(_ context.Context) (runtime.String, error) {
	return runtime.NewString(goquery.NodeName(el.selection)), nil
}

func (el *HTMLElement) Length(ctx context.Context) (runtime.Int, error) {
	if el.children == nil {
		el.children = el.parseChildren()
	}

	return el.children.Length(ctx)
}

func (el *HTMLElement) GetValue(_ context.Context) (runtime.Value, error) {
	val, ok := el.selection.Attr("value")

	if ok {
		return runtime.NewString(val), nil
	}

	return runtime.EmptyString, nil
}

func (el *HTMLElement) SetValue(_ context.Context, value runtime.Value) error {
	el.selection.SetAttr("value", value.String())

	return nil
}

func (el *HTMLElement) GetInnerText(_ context.Context) (runtime.String, error) {
	return runtime.NewString(el.selection.Text()), nil
}

func (el *HTMLElement) SetInnerText(_ context.Context, innerText runtime.String) error {
	el.selection.SetText(innerText.String())

	return nil
}

func (el *HTMLElement) GetInnerHTML(_ context.Context) (runtime.String, error) {
	h, err := el.selection.Html()

	if err != nil {
		return runtime.EmptyString, err
	}

	return runtime.NewString(h), nil
}

func (el *HTMLElement) SetInnerHTML(_ context.Context, value runtime.String) error {
	el.selection.SetHtml(value.String())

	return nil
}

func (el *HTMLElement) GetStyles(ctx context.Context) (runtime.Map, error) {
	if err := el.ensureStyles(ctx); err != nil {
		return runtime.NewObject(), err
	}

	return el.styles.Copy().(*runtime.Object), nil
}

func (el *HTMLElement) GetStyle(ctx context.Context, name runtime.String) (runtime.Value, error) {
	if err := el.ensureStyles(ctx); err != nil {
		return runtime.None, err
	}

	return el.styles.Get(ctx, name)
}

func (el *HTMLElement) SetStyle(ctx context.Context, name, value runtime.String) error {
	if err := el.ensureStyles(ctx); err != nil {
		return err
	}

	_ = el.styles.Set(ctx, name, value)

	str, err := styleutil.Serialize(ctx, el.styles)

	if err != nil {
		return err
	}

	return el.SetAttribute(ctx, "style", str)
}

func (el *HTMLElement) SetStyles(ctx context.Context, newStyles runtime.Map) error {
	if newStyles == nil {
		return nil
	}

	if err := el.ensureStyles(ctx); err != nil {
		return err
	}

	_ = newStyles.ForEach(ctx, func(ctx context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		_ = el.styles.Set(ctx, key, value)

		return true, nil
	})

	str, err := styleutil.Serialize(ctx, el.styles)

	if err != nil {
		return err
	}

	return el.SetAttribute(ctx, "style", str)
}

func (el *HTMLElement) RemoveStyle(ctx context.Context, name ...runtime.String) error {
	if len(name) == 0 {
		return nil
	}

	if err := el.ensureStyles(ctx); err != nil {
		return err
	}

	for _, s := range name {
		_ = el.styles.Remove(ctx, s)
	}

	str, err := styleutil.Serialize(ctx, el.styles)

	if err != nil {
		return err
	}

	return el.SetAttribute(ctx, "style", str)
}

func (el *HTMLElement) SetAttributes(ctx context.Context, attrs runtime.Map) error {
	if attrs == nil {
		return nil
	}

	el.ensureAttrs()

	return attrs.ForEach(ctx, func(ctx context.Context, key, value runtime.Value) (runtime.Boolean, error) {
		err := el.SetAttribute(ctx, runtime.NewString(key.String()), runtime.NewString(value.String()))

		return true, err
	})
}

func (el *HTMLElement) GetAttributes(ctx context.Context) (runtime.Map, error) {
	el.ensureAttrs()

	return el.attrs.Copy().(*runtime.Object), nil
}

func (el *HTMLElement) GetAttribute(ctx context.Context, name runtime.String) (runtime.Value, error) {
	el.ensureAttrs()

	if name == styleutil.AttributeNameStyle {
		return el.GetStyles(ctx)
	}

	exists, _ := el.attrs.ContainsKey(ctx, name)

	if !exists {
		return runtime.None, nil
	}

	return el.attrs.Get(ctx, name)
}

func (el *HTMLElement) SetAttribute(ctx context.Context, name, value runtime.String) error {
	el.ensureAttrs()

	if name == styleutil.AttributeNameStyle {
		el.styles = nil
	}

	_ = el.attrs.Set(ctx, name, value)
	el.selection.SetAttr(string(name), string(value))

	return nil
}

func (el *HTMLElement) RemoveAttribute(ctx context.Context, name ...runtime.String) error {
	el.ensureAttrs()

	for _, attr := range name {
		_ = el.attrs.Remove(ctx, attr)
		el.selection.RemoveAttr(attr.String())
	}

	return nil
}

func (el *HTMLElement) GetChildNodes(_ context.Context) (runtime.List, error) {
	if el.children == nil {
		el.children = el.parseChildren()
	}

	return el.children.Copy().(runtime.List), nil
}

func (el *HTMLElement) GetChildNode(ctx context.Context, idx runtime.Int) (runtime.Value, error) {
	if el.children == nil {
		el.children = el.parseChildren()
	}

	return el.children.At(ctx, idx)
}

func (el *HTMLElement) QuerySelector(_ context.Context, selector drivers.QuerySelector) (runtime.Value, error) {
	if selector.Kind == drivers.CSSSelector {
		selection := el.selection.Find(selector.String())

		if selection.Length() == 0 {
			return runtime.None, nil
		}

		res, err := NewHTMLElement(el.doc, selection)

		if err != nil {
			return runtime.None, err
		}

		return res, nil
	}

	found, err := EvalXPathToNode(el.doc, el.selection, selector.String())

	if err != nil {
		return runtime.None, err
	}

	if found == nil {
		return runtime.None, nil
	}

	return found, nil
}

func (el *HTMLElement) QuerySelectorAll(ctx context.Context, selector drivers.QuerySelector) (runtime.List, error) {
	if selector.Kind == drivers.CSSSelector {
		selection := el.selection.Find(selector.String())

		if selection.Length() == 0 {
			return runtime.NewArray(0), nil
		}

		arr := runtime.NewArray(selection.Length())

		selection.Each(func(_ int, selection *goquery.Selection) {
			el, err := NewHTMLElement(el.doc, selection)

			if err == nil {
				_ = arr.Append(ctx, el)
			}
		})

		return arr, nil
	}

	return EvalXPathToNodes(el.doc, el.selection, selector.String())
}

func (el *HTMLElement) XPath(_ context.Context, expression runtime.String) (runtime.Value, error) {
	return EvalXPathTo(el.doc, el.selection, expression.String())
}

func (el *HTMLElement) SetInnerHTMLBySelector(ctx context.Context, selector drivers.QuerySelector, innerHTML runtime.String) error {
	if selector.Kind == drivers.CSSSelector {
		selection := el.selection.Find(selector.String())

		if selection.Length() == 0 {
			return drivers.ErrNotFound
		}

		selection.SetHtml(innerHTML.String())
	}

	found, err := EvalXPathToElement(el.doc, el.selection, selector.String())

	if err != nil {
		return err
	}

	if found == nil {
		return drivers.ErrNotFound
	}

	return found.SetInnerHTML(ctx, innerHTML)
}

func (el *HTMLElement) GetInnerHTMLBySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.String, error) {
	if selector.Kind == drivers.CSSSelector {
		selection := el.selection.Find(selector.String())

		if selection.Length() == 0 {
			return runtime.EmptyString, drivers.ErrNotFound
		}

		str, err := selection.Html()

		if err != nil {
			return runtime.EmptyString, err
		}

		return runtime.NewString(str), nil
	}

	found, err := EvalXPathToElement(el.doc, el.selection, selector.String())

	if err != nil {
		return runtime.EmptyString, err
	}

	if found == nil {
		return runtime.EmptyString, drivers.ErrNotFound
	}

	return found.GetInnerHTML(ctx)
}

func (el *HTMLElement) GetInnerHTMLBySelectorAll(ctx context.Context, selector drivers.QuerySelector) (runtime.List, error) {
	if selector.Kind == drivers.CSSSelector {
		var err error
		selection := el.selection.Find(selector.String())
		arr := runtime.NewArray(selection.Length())

		selection.EachWithBreak(func(_ int, selection *goquery.Selection) bool {
			str, e := selection.Html()

			if e != nil {
				err = e
				return false
			}

			_ = arr.Append(ctx, runtime.NewString(strings.TrimSpace(str)))

			return true
		})

		if err != nil {
			return runtime.NewArray(0), err
		}

		return arr, nil
	}

	return EvalXPathToNodesWith(el.selection, selector.String(), func(node *html.Node) (runtime.Value, error) {
		n, err := parseXPathNode(el.doc, node)

		if err != nil {
			return runtime.None, err
		}

		found, err := drivers.ToElement(n)

		if err != nil {
			return runtime.None, err
		}

		return found.GetInnerHTML(ctx)
	})
}

func (el *HTMLElement) GetInnerTextBySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.String, error) {
	if selector.Kind == drivers.CSSSelector {
		selection := el.selection.Find(selector.String())

		if selection.Length() == 0 {
			return runtime.EmptyString, drivers.ErrNotFound
		}

		return runtime.NewString(selection.Text()), nil
	}

	found, err := EvalXPathToElement(el.doc, el.selection, selector.String())

	if err != nil {
		return runtime.EmptyString, err
	}

	if found == nil {
		return runtime.EmptyString, drivers.ErrNotFound
	}

	return found.GetInnerText(ctx)
}

func (el *HTMLElement) SetInnerTextBySelector(ctx context.Context, selector drivers.QuerySelector, innerText runtime.String) error {
	if selector.Kind == drivers.CSSSelector {
		selection := el.selection.Find(selector.String())

		if selection.Length() == 0 {
			return drivers.ErrNotFound
		}

		selection.SetHtml(innerText.String())

		return nil
	}

	found, err := EvalXPathToElement(el.doc, el.selection, selector.String())

	if err != nil {
		return err
	}

	if found == nil {
		return drivers.ErrNotFound
	}

	return found.SetInnerText(ctx, innerText)
}

func (el *HTMLElement) GetInnerTextBySelectorAll(ctx context.Context, selector drivers.QuerySelector) (runtime.List, error) {
	if selector.Kind == drivers.CSSSelector {
		selection := el.selection.Find(selector.String())
		arr := runtime.NewArray(selection.Length())

		selection.Each(func(_ int, selection *goquery.Selection) {
			_ = arr.Append(ctx, runtime.NewString(selection.Text()))
		})

		return arr, nil
	}

	return EvalXPathToNodesWith(el.selection, selector.String(), func(node *html.Node) (runtime.Value, error) {
		n, err := parseXPathNode(el.doc, node)

		if err != nil {
			return runtime.None, err
		}

		found, err := drivers.ToElement(n)

		if err != nil {
			return runtime.None, err
		}

		return found.GetInnerText(ctx)
	})
}

func (el *HTMLElement) CountBySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.Int, error) {
	if selector.Kind == drivers.CSSSelector {
		selection := el.selection.Find(selector.String())

		return runtime.NewInt(selection.Length()), nil
	}

	arr, err := EvalXPathToNodesWith(el.selection, selector.String(), func(_ *html.Node) (runtime.Value, error) {
		return runtime.None, nil
	})

	if err != nil {
		return runtime.ZeroInt, err
	}

	return arr.Length(ctx)
}

func (el *HTMLElement) ExistsBySelector(_ context.Context, selector drivers.QuerySelector) (runtime.Boolean, error) {
	if selector.Kind == drivers.CSSSelector {
		selection := el.selection.Find(selector.String())

		if selection.Length() == 0 {
			return runtime.False, nil
		}

		return runtime.True, nil
	}

	found, err := EvalXPathToNode(el.doc, el.selection, selector.String())

	if err != nil {
		return runtime.False, err
	}

	return runtime.NewBoolean(found != nil), nil
}

func (el *HTMLElement) Get(ctx context.Context, path runtime.Value) (runtime.Value, error) {
	return access.GetInElement(ctx, path, el)
}

func (el *HTMLElement) Iterate(_ context.Context) (runtime.Iterator, error) {
	return access.NewIterator(el)
}

func (el *HTMLElement) GetParentElement(_ context.Context) (runtime.Value, error) {
	parent := el.selection.Parent()

	if parent == nil {
		return runtime.None, nil
	}

	return NewHTMLElement(el.doc, parent)
}

func (el *HTMLElement) GetPreviousElementSibling(_ context.Context) (runtime.Value, error) {
	sibling := el.selection.Prev()

	if sibling == nil {
		return runtime.None, nil
	}

	return NewHTMLElement(el.doc, sibling)
}

func (el *HTMLElement) GetNextElementSibling(_ context.Context) (runtime.Value, error) {
	sibling := el.selection.Next()

	if sibling == nil {
		return runtime.None, nil
	}

	return NewHTMLElement(el.doc, sibling)
}

func (el *HTMLElement) Query(ctx context.Context, q runtime.Query) (runtime.List, error) {
	switch queryutil.Parse(string(q.Kind)) {
	case queryutil.CSS:
		return EvalCSSX(ctx, el, q.Payload)
	case queryutil.XPath:
		res, err := el.XPath(ctx, q.Payload)

		if err != nil {
			return nil, err
		}

		list, ok := res.(runtime.List)

		if !ok {
			return runtime.NewArrayWith(res), nil
		}

		return list, nil
	default:
		return nil, runtime.Error(runtime.ErrInvalidArgument, "unsupported query kind")
	}
}

func (el *HTMLElement) ensureStyles(ctx context.Context) error {
	if el.styles == nil {
		styles, err := el.parseStyles(ctx)

		if err != nil {
			return err
		}

		el.styles = styles
	}

	return nil
}

func (el *HTMLElement) parseStyles(ctx context.Context) (*runtime.Object, error) {
	el.ensureAttrs()

	str, err := el.attrs.Get(ctx, runtime.NewString(styleutil.AttributeNameStyle))

	if err != nil {
		return runtime.NewObject(), err
	}

	if str == runtime.None {
		return runtime.NewObject(), nil
	}

	styles, err := styleutil.Deserialize(ctx, runtime.NewString(str.String()))

	if err != nil {
		return nil, err
	}

	return styles, nil
}

func (el *HTMLElement) ensureAttrs() {
	if el.attrs == nil {
		el.attrs = el.parseAttrs()
	}
}

func (el *HTMLElement) parseAttrs() *runtime.Object {
	obj := runtime.NewObject()

	if len(el.selection.Nodes) == 0 {
		return obj
	}

	node := el.selection.Nodes[0]
	ctx := context.Background()

	for _, attr := range node.Attr {
		_ = obj.Set(ctx, runtime.NewString(attr.Key), runtime.NewString(attr.Val))
	}

	return obj
}

func (el *HTMLElement) parseChildren() *runtime.Array {
	children := el.selection.Children()

	arr := runtime.NewArray(10)
	ctx := context.Background()

	children.Each(func(_ int, selection *goquery.Selection) {
		child, err := NewHTMLElement(el.doc, selection)

		if err == nil {
			_ = arr.Append(ctx, child)
		}
	})

	return arr
}
