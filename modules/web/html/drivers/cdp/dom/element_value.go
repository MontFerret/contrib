package dom

import (
	"context"
	"hash/fnv"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/contrib/modules/web/html/drivers/internal/data"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (el *HTMLElement) Type() runtime.Type {
	return drivers.HTMLElementType
}

func (el *HTMLElement) MarshalJSON() ([]byte, error) {
	if err := el.ensureAttached(); err != nil {
		return nil, err
	}

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

func (el *HTMLElement) Equal(ctx context.Context, other runtime.Value) (bool, error) {
	if _, err := checkComparisonContext(ctx); err != nil {
		return false, err
	}

	otherElement, ok := other.(*HTMLElement)
	if !ok {
		return false, nil
	}

	comparison, err := el.compare(ctx, otherElement)

	return comparison == runtime.Equal, err
}

func (el *HTMLElement) Compare(ctx context.Context, other runtime.Value) (runtime.Ordering, error) {
	if comparison, err := checkComparisonContext(ctx); err != nil {
		return comparison, err
	}

	otherElement, ok := other.(*HTMLElement)
	if ok {
		return el.compare(ctx, otherElement)
	}

	if _, ok := other.(drivers.HTMLElement); ok {
		if comparison, err := checkComparisonContext(ctx); err != nil {
			return comparison, err
		}

		return runtime.Greater, nil
	}

	return invalidComparison(el, other)
}

func (el *HTMLElement) compare(ctx context.Context, other *HTMLElement) (runtime.Ordering, error) {
	if comparison, err := checkComparisonContext(ctx); err != nil {
		return comparison, err
	}

	return runtime.Ordering(strings.Compare(string(el.id), string(other.id))), nil
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
	if err := el.ensureAttached(); err != nil {
		return nil, err
	}

	return data.NewIterator(el)
}

func (el *HTMLElement) Get(ctx context.Context, key runtime.Value) (runtime.Value, error) {
	if el.document != nil {
		if err := el.ensureAttached(); err != nil {
			return runtime.None, err
		}
	}

	if key == nil || key == runtime.None {
		return runtime.None, nil
	}

	switch key.String() {
	case "textContent":
		return el.GetTextContent(ctx)
	case "innerText":
		return el.GetInnerText(ctx)
	case "innerHTML":
		return el.GetInnerHTML(ctx)
	case "checked", "disabled", "selected":
		return el.eval.EvalValue(ctx, templates.GetDOMProperty(el.id, runtime.ToString(key)))
	case "attributes":
		return newAttributeView(ctx, el.attributes)
	case "style":
		return newStyleView(ctx, el.styles)
	case "classes":
		return newClassListView(ctx, el.classes)
	case "dataset":
		return newDatasetView(ctx, el.dataset)
	}

	return data.GetInElement(ctx, key, el)
}

func (el *HTMLElement) Set(ctx context.Context, key, value runtime.Value) error {
	if el.document != nil {
		if err := el.ensureAttached(); err != nil {
			return err
		}
	}

	if key == nil || key == runtime.None {
		return runtime.Error(runtime.ErrInvalidArgument, "element property name is empty")
	}

	if value == nil {
		value = runtime.None
	}

	name := runtime.ToString(key)

	switch name {
	case "textContent":
		return el.SetTextContent(ctx, runtime.ToString(value))
	case "innerText":
		return el.SetInnerText(ctx, runtime.ToString(value))
	case "innerHTML":
		return el.SetInnerHTML(ctx, runtime.ToString(value))
	case "value":
		return el.SetValue(ctx, runtime.ToString(value))
	case "checked", "disabled", "selected":
		enabled, err := runtime.CastBoolean(value)
		if err != nil {
			return err
		}

		return el.eval.Eval(ctx, templates.SetDOMProperty(el.id, name, enabled))
	case "attributes":
		attrs, err := runtime.CastMap(value)
		if err != nil {
			return err
		}

		return el.attributes.SetAttributes(ctx, attrs)
	case "style":
		styles, err := runtime.CastMap(value)
		if err != nil {
			return err
		}

		return el.styles.SetStyles(ctx, styles)
	case "classes":
		classes, err := runtime.CastArray(value)
		if err != nil {
			return err
		}

		return el.classes.SetClasses(ctx, classes)
	case "dataset":
		dataset, err := runtime.CastMap(value)
		if err != nil {
			return err
		}

		return el.dataset.SetDataset(ctx, dataset)
	default:
		return runtime.Errorf(runtime.ErrInvalidArgument, "element property %q is not writable", name)
	}
}

func (el *HTMLElement) GetValue(ctx context.Context) (runtime.Value, error) {
	return el.eval.EvalValue(ctx, templates.GetValue(el.id))
}

func (el *HTMLElement) GetDOMProperty(ctx context.Context, name runtime.String) (runtime.Value, error) {
	if el.eval == nil {
		return runtime.None, nil
	}

	return el.eval.EvalResult(ctx, templates.GetDOMProperty(el.id, name))
}

func (el *HTMLElement) GetNodeType(ctx context.Context) (runtime.Int, error) {
	if err := el.ensureAttached(); err != nil {
		return runtime.ZeroInt, err
	}

	out, err := el.nodeType.Read(ctx)
	if err != nil {
		return runtime.ZeroInt, err
	}

	return runtime.ToInt(ctx, out)
}

func (el *HTMLElement) GetNodeName(ctx context.Context) (runtime.String, error) {
	if err := el.ensureAttached(); err != nil {
		return runtime.EmptyString, err
	}

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
