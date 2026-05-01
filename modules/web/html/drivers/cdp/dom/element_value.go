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
	return data.NewIterator(el)
}

func (el *HTMLElement) Get(ctx context.Context, key runtime.Value) (runtime.Value, error) {
	return data.GetInElement(ctx, key, el)
}

func (el *HTMLElement) GetValue(ctx context.Context) (runtime.Value, error) {
	return el.eval.EvalValue(ctx, templates.GetValue(el.id))
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
