package dom

import (
	"context"
	"hash/fnv"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/contrib/modules/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/html/drivers/cdp/events"
	"github.com/MontFerret/contrib/modules/html/drivers/cdp/input"
	"github.com/MontFerret/contrib/modules/html/drivers/cdp/templates"
	"github.com/MontFerret/contrib/modules/html/drivers/common"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type HTMLDocument struct {
	logger    zerolog.Logger
	client    *cdp.Client
	dom       *Manager
	input     *input.Manager
	eval      *eval.Runtime
	frameTree page.FrameTree
	element   *HTMLElement
}

func NewHTMLDocument(
	logger zerolog.Logger,
	client *cdp.Client,
	domManager *Manager,
	input *input.Manager,
	exec *eval.Runtime,
	rootElement *HTMLElement,
	frames page.FrameTree,
) *HTMLDocument {
	doc := new(HTMLDocument)
	doc.logger = common.LoggerWithName(logger.With(), "html_document").Logger()
	doc.client = client
	doc.dom = domManager
	doc.input = input
	doc.eval = exec
	doc.element = rootElement
	doc.frameTree = frames

	return doc
}

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
	return common.GetInDocument(ctx, key, doc)
}

func (doc *HTMLDocument) Set(ctx context.Context, key runtime.Value, value runtime.Value) error {
	return common.SetInDocument(ctx, key, doc, value)
}

func (doc *HTMLDocument) Close() error {
	return doc.element.Close()
}

func (doc *HTMLDocument) Frame() page.FrameTree {
	return doc.frameTree
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

func (doc *HTMLDocument) GetParentDocument(ctx context.Context) (drivers.HTMLDocument, error) {
	if doc.frameTree.Frame.ParentID == nil {
		return nil, nil
	}

	return doc.dom.GetFrameNode(ctx, *doc.frameTree.Frame.ParentID)
}

func (doc *HTMLDocument) GetChildDocuments(ctx context.Context) (runtime.List, error) {
	arr := runtime.NewArray(len(doc.frameTree.ChildFrames))

	for _, childFrame := range doc.frameTree.ChildFrames {
		frame, err := doc.dom.GetFrameNode(ctx, childFrame.Frame.ID)

		if err != nil {
			return nil, err
		}

		if frame != nil {
			_ = arr.Append(ctx, frame)
		}
	}

	return arr, nil
}

func (doc *HTMLDocument) XPath(ctx context.Context, expression runtime.String) (runtime.Value, error) {
	return doc.element.XPath(ctx, expression)
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

func (doc *HTMLDocument) MoveMouseByXY(ctx context.Context, x, y runtime.Float) error {
	return doc.input.MoveMouseByXY(ctx, x, y)
}

func (doc *HTMLDocument) WaitForElement(ctx context.Context, selector drivers.QuerySelector, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		doc.eval,
		templates.WaitForElement(doc.element.id, selector, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (doc *HTMLDocument) WaitForClassBySelector(ctx context.Context, selector drivers.QuerySelector, class runtime.String, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		doc.eval,
		templates.WaitForClassBySelector(doc.element.id, selector, class, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (doc *HTMLDocument) WaitForClassBySelectorAll(ctx context.Context, selector drivers.QuerySelector, class runtime.String, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		doc.eval,
		templates.WaitForClassBySelectorAll(doc.element.id, selector, class, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (doc *HTMLDocument) WaitForAttributeBySelector(
	ctx context.Context,
	selector drivers.QuerySelector,
	name,
	value runtime.String,
	when drivers.WaitEvent,
) error {
	task := events.NewEvalWaitTask(
		doc.eval,
		templates.WaitForAttributeBySelector(doc.element.id, selector, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (doc *HTMLDocument) WaitForAttributeBySelectorAll(
	ctx context.Context,
	selector drivers.QuerySelector,
	name,
	value runtime.String,
	when drivers.WaitEvent,
) error {
	task := events.NewEvalWaitTask(
		doc.eval,
		templates.WaitForAttributeBySelectorAll(doc.element.id, selector, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (doc *HTMLDocument) WaitForStyleBySelector(ctx context.Context, selector drivers.QuerySelector, name, value runtime.String, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		doc.eval,
		templates.WaitForStyleBySelector(doc.element.id, selector, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (doc *HTMLDocument) WaitForStyleBySelectorAll(ctx context.Context, selector drivers.QuerySelector, name, value runtime.String, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		doc.eval,
		templates.WaitForStyleBySelectorAll(doc.element.id, selector, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (doc *HTMLDocument) ScrollTop(ctx context.Context, options drivers.ScrollOptions) error {
	return doc.input.ScrollTop(ctx, options)
}

func (doc *HTMLDocument) ScrollBottom(ctx context.Context, options drivers.ScrollOptions) error {
	return doc.input.ScrollBottom(ctx, options)
}

func (doc *HTMLDocument) ScrollBySelector(ctx context.Context, selector drivers.QuerySelector, options drivers.ScrollOptions) error {
	return doc.input.ScrollIntoViewBySelector(ctx, doc.element.id, selector, options)
}

func (doc *HTMLDocument) Scroll(ctx context.Context, options drivers.ScrollOptions) error {
	return doc.input.ScrollByXY(ctx, options)
}

func (doc *HTMLDocument) Eval() *eval.Runtime {
	return doc.eval
}

func (doc *HTMLDocument) Query(ctx context.Context, q runtime.Query) (runtime.Value, error) {
	//TODO implement me
	panic("implement me")
}

func (doc *HTMLDocument) Dispatch(ctx context.Context, event runtime.DispatchEvent) (runtime.Value, error) {
	//TODO implement me
	panic("implement me")
}

func (doc *HTMLDocument) logError(err error) *zerolog.Event {
	return doc.logger.
		Error().
		Timestamp().
		Str("url", doc.frameTree.Frame.URL).
		Str("securityOrigin", doc.frameTree.Frame.SecurityOrigin).
		Str("mimeType", doc.frameTree.Frame.MimeType).
		Str("frameID", string(doc.frameTree.Frame.ID)).
		Err(err)
}
