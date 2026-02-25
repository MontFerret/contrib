package dom

import (
	"context"
	"hash/fnv"
	"strings"
	"time"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/contrib/modules/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/html/drivers/cdp/events"
	"github.com/MontFerret/contrib/modules/html/drivers/cdp/input"
	"github.com/MontFerret/contrib/modules/html/drivers/cdp/templates"
	"github.com/MontFerret/contrib/modules/html/drivers/common"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
	"github.com/goccy/go-json"
	"github.com/mafredri/cdp"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type HTMLElement struct {
	logger   zerolog.Logger
	client   *cdp.Client
	dom      *Manager
	input    *input.Manager
	eval     *eval.Runtime
	id       cdpruntime.RemoteObjectID
	nodeType *common.LazyValue
	nodeName *common.LazyValue
}

func NewHTMLElement(
	logger zerolog.Logger,
	client *cdp.Client,
	domManager *Manager,
	input *input.Manager,
	exec *eval.Runtime,
	id cdpruntime.RemoteObjectID,
) *HTMLElement {
	el := new(HTMLElement)
	el.logger = common.
		LoggerWithName(logger.With(), "dom_element").
		Str("object_id", string(id)).
		Logger()
	el.client = client
	el.dom = domManager
	el.input = input
	el.eval = exec
	el.id = id
	el.nodeType = common.NewLazyValue(func(ctx context.Context) (runtime.Value, error) {
		return el.eval.EvalValue(ctx, templates.GetNodeType(el.id))
	})
	el.nodeName = common.NewLazyValue(func(ctx context.Context) (runtime.Value, error) {
		return el.eval.EvalValue(ctx, templates.GetNodeName(el.id))
	})

	return el
}

func (el *HTMLElement) RemoteID() cdpruntime.RemoteObjectID {
	return el.id
}

func (el *HTMLElement) Close() error {
	return nil
}

func (el *HTMLElement) Type() runtime.Type {
	return drivers.HTMLElementType
}

func (el *HTMLElement) MarshalJSON() ([]byte, error) {
	return json.Marshal(el.String())
}

func (el *HTMLElement) String() string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(drivers.DefaultWaitTimeout)*time.Millisecond)
	defer cancel()

	res, err := el.GetInnerHTML(ctx)

	if err != nil {
		el.logError(errors.Wrap(err, "HTMLElement.String"))

		return ""
	}

	return res.String()
}

func (el *HTMLElement) Compare(other runtime.Value) int {
	cdpEl, ok := other.(*HTMLElement)

	if ok {
		return strings.Compare(string(el.id), string(cdpEl.id))
	}

	genericEl, ok := other.(drivers.HTMLElement)

	if ok {
		return strings.Compare(el.String(), genericEl.String())
	}

	return drivers.CompareTypes(el, other)
}

func (el *HTMLElement) Unwrap() any {
	return el
}

func (el *HTMLElement) Hash() uint64 {
	h := fnv.New64a()

	h.Write([]byte(el.Type().String()))
	h.Write([]byte(":"))
	h.Write([]byte(el.id))

	return h.Sum64()
}

func (el *HTMLElement) Copy() runtime.Value {
	return runtime.None
}

func (el *HTMLElement) Iterate(_ context.Context) (runtime.Iterator, error) {
	return common.NewIterator(el)
}

func (el *HTMLElement) Get(ctx context.Context, key runtime.Value) (runtime.Value, error) {
	return common.GetInElement(ctx, key, el)
}

func (el *HTMLElement) Set(ctx context.Context, key, value runtime.Value) error {
	return common.SetInElement(ctx, key, el, value)
}

func (el *HTMLElement) GetValue(ctx context.Context) (runtime.Value, error) {
	return el.eval.EvalValue(ctx, templates.GetValue(el.id))
}

func (el *HTMLElement) SetValue(ctx context.Context, value runtime.Value) error {
	return el.eval.Eval(ctx, templates.SetValue(el.id, value))
}

func (el *HTMLElement) GetNodeType(ctx context.Context) (runtime.Int, error) {
	out, err := el.nodeType.Read(ctx)

	if err != nil {
		return runtime.ZeroInt, err
	}

	return runtime.ToInt(ctx, out)
}

func (el *HTMLElement) GetNodeName(ctx context.Context) (runtime.String, error) {
	out, err := el.nodeName.Read(ctx)

	if err != nil {
		return runtime.EmptyString, err
	}

	return runtime.ToString(out), nil
}

func (el *HTMLElement) Length(ctx context.Context) (runtime.Int, error) {
	value, err := el.eval.EvalValue(ctx, templates.GetChildrenCount(el.id))

	if err != nil {
		el.logError(err)

		return 0, errors.Wrap(err, "failed to get children count")
	}

	return runtime.ToInt(ctx, value)
}

func (el *HTMLElement) GetStyles(ctx context.Context) (runtime.Map, error) {
	out, err := el.eval.EvalValue(ctx, templates.GetStyles(el.id))

	if err != nil {
		return runtime.NewObject(), err
	}

	return runtime.ToMap(ctx, out)
}

func (el *HTMLElement) GetStyle(ctx context.Context, name runtime.String) (runtime.Value, error) {
	return el.eval.EvalValue(ctx, templates.GetStyle(el.id, name))
}

func (el *HTMLElement) SetStyles(ctx context.Context, styles runtime.Map) error {
	return el.eval.Eval(ctx, templates.SetStyles(el.id, styles))
}

func (el *HTMLElement) SetStyle(ctx context.Context, name, value runtime.String) error {
	return el.eval.Eval(ctx, templates.SetStyle(el.id, name, value))
}

func (el *HTMLElement) RemoveStyle(ctx context.Context, names ...runtime.String) error {
	return el.eval.Eval(ctx, templates.RemoveStyles(el.id, names))
}

func (el *HTMLElement) GetAttributes(ctx context.Context) (runtime.Map, error) {
	out, err := el.eval.EvalValue(ctx, templates.GetAttributes(el.id))

	if err != nil {
		return runtime.NewObject(), err
	}

	return runtime.ToMap(ctx, out)
}

func (el *HTMLElement) GetAttribute(ctx context.Context, name runtime.String) (runtime.Value, error) {
	return el.eval.EvalValue(ctx, templates.GetAttribute(el.id, name))
}

func (el *HTMLElement) SetAttributes(ctx context.Context, attrs runtime.Map) error {
	return el.eval.Eval(ctx, templates.SetAttributes(el.id, attrs))
}

func (el *HTMLElement) SetAttribute(ctx context.Context, name, value runtime.String) error {
	return el.eval.Eval(ctx, templates.SetAttribute(el.id, name, value))
}

func (el *HTMLElement) RemoveAttribute(ctx context.Context, names ...runtime.String) error {
	return el.eval.Eval(ctx, templates.RemoveAttributes(el.id, names))
}

func (el *HTMLElement) GetChildNodes(ctx context.Context) (runtime.List, error) {
	return el.eval.EvalElements(ctx, templates.GetChildren(el.id))
}

func (el *HTMLElement) GetChildNode(ctx context.Context, idx runtime.Int) (runtime.Value, error) {
	return el.eval.EvalElement(ctx, templates.GetChildByIndex(el.id, idx))
}

func (el *HTMLElement) GetParentElement(ctx context.Context) (runtime.Value, error) {
	return el.eval.EvalElement(ctx, templates.GetParent(el.id))
}

func (el *HTMLElement) GetPreviousElementSibling(ctx context.Context) (runtime.Value, error) {
	return el.eval.EvalElement(ctx, templates.GetPreviousElementSibling(el.id))
}

func (el *HTMLElement) GetNextElementSibling(ctx context.Context) (runtime.Value, error) {
	return el.eval.EvalElement(ctx, templates.GetNextElementSibling(el.id))
}

func (el *HTMLElement) QuerySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.Value, error) {
	return el.eval.EvalElement(ctx, templates.QuerySelector(el.id, selector))
}

func (el *HTMLElement) QuerySelectorAll(ctx context.Context, selector drivers.QuerySelector) (runtime.List, error) {
	return el.eval.EvalElements(ctx, templates.QuerySelectorAll(el.id, selector))
}

func (el *HTMLElement) XPath(ctx context.Context, expression runtime.String) (result runtime.Value, err error) {
	return el.eval.EvalValue(ctx, templates.XPath(el.id, expression))
}

func (el *HTMLElement) GetInnerText(ctx context.Context) (runtime.String, error) {
	out, err := el.eval.EvalValue(ctx, templates.GetInnerText(el.id))

	if err != nil {
		return runtime.EmptyString, err
	}

	return runtime.ToString(out), nil
}

func (el *HTMLElement) SetInnerText(ctx context.Context, innerText runtime.String) error {
	return el.eval.Eval(
		ctx,
		templates.SetInnerText(el.id, innerText),
	)
}

func (el *HTMLElement) GetInnerTextBySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.String, error) {
	out, err := el.eval.EvalValue(ctx, templates.GetInnerTextBySelector(el.id, selector))

	if err != nil {
		return runtime.EmptyString, err
	}

	return runtime.ToString(out), nil
}

func (el *HTMLElement) SetInnerTextBySelector(ctx context.Context, selector drivers.QuerySelector, innerText runtime.String) error {
	return el.eval.Eval(
		ctx,
		templates.SetInnerTextBySelector(el.id, selector, innerText),
	)
}

func (el *HTMLElement) GetInnerTextBySelectorAll(ctx context.Context, selector drivers.QuerySelector) (runtime.List, error) {
	out, err := el.eval.EvalValue(ctx, templates.GetInnerTextBySelectorAll(el.id, selector))

	if err != nil {
		return runtime.EmptyArray(), err
	}

	return runtime.ToList(ctx, out)
}

func (el *HTMLElement) GetInnerHTML(ctx context.Context) (runtime.String, error) {
	out, err := el.eval.EvalValue(ctx, templates.GetInnerHTML(el.id))

	if err != nil {
		return runtime.EmptyString, err
	}

	return runtime.ToString(out), nil
}

func (el *HTMLElement) SetInnerHTML(ctx context.Context, innerHTML runtime.String) error {
	return el.eval.Eval(ctx, templates.SetInnerHTML(el.id, innerHTML))
}

func (el *HTMLElement) GetInnerHTMLBySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.String, error) {
	out, err := el.eval.EvalValue(ctx, templates.GetInnerHTMLBySelector(el.id, selector))

	if err != nil {
		return runtime.EmptyString, err
	}

	return runtime.ToString(out), nil
}

func (el *HTMLElement) SetInnerHTMLBySelector(ctx context.Context, selector drivers.QuerySelector, innerHTML runtime.String) error {
	return el.eval.Eval(ctx, templates.SetInnerHTMLBySelector(el.id, selector, innerHTML))
}

func (el *HTMLElement) GetInnerHTMLBySelectorAll(ctx context.Context, selector drivers.QuerySelector) (runtime.List, error) {
	out, err := el.eval.EvalValue(ctx, templates.GetInnerHTMLBySelectorAll(el.id, selector))

	if err != nil {
		return runtime.EmptyArray(), err
	}

	return runtime.ToList(ctx, out)
}

func (el *HTMLElement) CountBySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.Int, error) {
	out, err := el.eval.EvalValue(ctx, templates.CountBySelector(el.id, selector))

	if err != nil {
		return runtime.ZeroInt, err
	}

	return runtime.ToInt(ctx, out)
}

func (el *HTMLElement) ExistsBySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.Boolean, error) {
	out, err := el.eval.EvalValue(ctx, templates.ExistsBySelector(el.id, selector))

	if err != nil {
		return runtime.False, err
	}

	return runtime.ToBoolean(out), nil
}

func (el *HTMLElement) WaitForElement(ctx context.Context, selector drivers.QuerySelector, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForElement(el.id, selector, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForElementAll(ctx context.Context, selector drivers.QuerySelector, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForElementAll(el.id, selector, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForClass(ctx context.Context, class runtime.String, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForClass(el.id, class, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForClassBySelector(ctx context.Context, selector drivers.QuerySelector, class runtime.String, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForClassBySelector(el.id, selector, class, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForClassBySelectorAll(ctx context.Context, selector drivers.QuerySelector, class runtime.String, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForClassBySelectorAll(el.id, selector, class, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForAttribute(
	ctx context.Context,
	name runtime.String,
	value runtime.Value,
	when drivers.WaitEvent,
) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForAttribute(el.id, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForAttributeBySelector(ctx context.Context, selector drivers.QuerySelector, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForAttributeBySelector(el.id, selector, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForAttributeBySelectorAll(ctx context.Context, selector drivers.QuerySelector, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForAttributeBySelectorAll(el.id, selector, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForStyle(ctx context.Context, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForStyle(el.id, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForStyleBySelector(ctx context.Context, selector drivers.QuerySelector, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForStyleBySelector(el.id, selector, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForStyleBySelectorAll(ctx context.Context, selector drivers.QuerySelector, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForStyleBySelectorAll(el.id, selector, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) Click(ctx context.Context, count runtime.Int) error {
	return el.input.Click(ctx, el.id, int(count))
}

func (el *HTMLElement) ClickBySelector(ctx context.Context, selector drivers.QuerySelector, count runtime.Int) error {
	return el.input.ClickBySelector(ctx, el.id, selector, count)
}

func (el *HTMLElement) ClickBySelectorAll(ctx context.Context, selector drivers.QuerySelector, count runtime.Int) error {
	elements, err := el.QuerySelectorAll(ctx, selector)

	if err != nil {
		return err
	}

	return elements.ForEach(ctx, func(ctx context.Context, value runtime.Value, idx runtime.Int) (runtime.Boolean, error) {
		found := value.(*HTMLElement)

		if e := found.Click(ctx, count); e != nil {
			err = e
			return false, e
		}

		return true, nil
	})
}

func (el *HTMLElement) Input(ctx context.Context, value runtime.Value, delay runtime.Int) error {
	name, err := el.GetNodeName(ctx)

	if err != nil {
		return err
	}

	if strings.ToLower(string(name)) != "input" {
		return runtime.Error(runtime.ErrInvalidOperation, "element is not an <input> element.")
	}

	return el.input.Type(ctx, el.id, input.TypeParams{
		Text:  value.String(),
		Clear: false,
		Delay: time.Duration(delay) * time.Millisecond,
	})
}

func (el *HTMLElement) InputBySelector(ctx context.Context, selector drivers.QuerySelector, value runtime.Value, delay runtime.Int) error {
	return el.input.TypeBySelector(ctx, el.id, selector, input.TypeParams{
		Text:  value.String(),
		Clear: false,
		Delay: time.Duration(delay) * time.Millisecond,
	})
}

func (el *HTMLElement) Press(ctx context.Context, keys []runtime.String, count runtime.Int) error {
	return el.input.Press(ctx, sdk.UnwrapStrings(keys), int(count))
}

func (el *HTMLElement) PressBySelector(ctx context.Context, selector drivers.QuerySelector, keys []runtime.String, count runtime.Int) error {
	return el.input.PressBySelector(ctx, el.id, selector, sdk.UnwrapStrings(keys), int(count))
}

func (el *HTMLElement) Clear(ctx context.Context) error {
	return el.input.Clear(ctx, el.id)
}

func (el *HTMLElement) ClearBySelector(ctx context.Context, selector drivers.QuerySelector) error {
	return el.input.ClearBySelector(ctx, el.id, selector)
}

func (el *HTMLElement) Select(ctx context.Context, value runtime.List) (runtime.List, error) {
	return el.input.Select(ctx, el.id, value)
}

func (el *HTMLElement) SelectBySelector(ctx context.Context, selector drivers.QuerySelector, value runtime.List) (runtime.List, error) {
	return el.input.SelectBySelector(ctx, el.id, selector, value)
}

func (el *HTMLElement) ScrollIntoView(ctx context.Context, options drivers.ScrollOptions) error {
	return el.input.ScrollIntoView(ctx, el.id, options)
}

func (el *HTMLElement) Focus(ctx context.Context) error {
	return el.input.Focus(ctx, el.id)
}

func (el *HTMLElement) FocusBySelector(ctx context.Context, selector drivers.QuerySelector) error {
	return el.input.FocusBySelector(ctx, el.id, selector)
}

func (el *HTMLElement) Blur(ctx context.Context) error {
	return el.input.Blur(ctx, el.id)
}

func (el *HTMLElement) BlurBySelector(ctx context.Context, selector drivers.QuerySelector) error {
	return el.input.BlurBySelector(ctx, el.id, selector)
}

func (el *HTMLElement) Hover(ctx context.Context) error {
	return el.input.MoveMouse(ctx, el.id)
}

func (el *HTMLElement) HoverBySelector(ctx context.Context, selector drivers.QuerySelector) error {
	return el.input.MoveMouseBySelector(ctx, el.id, selector)
}

func (el *HTMLElement) Query(ctx context.Context, q runtime.Query) (runtime.Value, error) {
	//TODO implement me
	panic("implement me")
}

func (el *HTMLElement) Dispatch(ctx context.Context, event runtime.DispatchEvent) (runtime.Value, error) {
	//TODO implement me
	panic("implement me")
}

func (el *HTMLElement) logError(err error) *zerolog.Event {
	return el.logger.
		Error().
		Timestamp().
		Str("objectID", string(el.id)).
		Err(err)
}
