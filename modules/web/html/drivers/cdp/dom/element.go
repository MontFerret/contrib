package dom

import (
	"context"

	"github.com/mafredri/cdp"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/input"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/contrib/modules/web/html/drivers/internal/lazy"
	"github.com/MontFerret/contrib/modules/web/html/internal/logutil"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type HTMLElement struct {
	logger     zerolog.Logger
	client     *cdp.Client
	dom        *Manager
	input      *input.Manager
	eval       *eval.Runtime
	attributes *elementAttributes
	styles     *elementStyles
	classes    *elementClasses
	dataset    *elementDataset
	wait       *elementWait
	nodeType   *lazy.Value
	nodeName   *lazy.Value
	id         cdpruntime.RemoteObjectID
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
	el.logger = logutil.WithComponent(logger.With(), "dom_element").
		Str("object_id", string(id)).
		Logger()
	el.client = client
	el.dom = domManager
	el.input = input
	el.eval = exec
	el.id = id
	el.attributes = newElementAttributes(exec, id)
	el.styles = newElementStyles(exec, id)
	el.classes = newElementClasses(exec, id)
	el.dataset = newElementDataset(exec, id)
	el.wait = newElementWait(exec, id)
	el.nodeType = lazy.New(func(ctx context.Context) (runtime.Value, error) {
		return el.eval.EvalValue(ctx, templates.GetNodeType(el.id))
	})
	el.nodeName = lazy.New(func(ctx context.Context) (runtime.Value, error) {
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

func (el *HTMLElement) AsContentTarget() drivers.ContentTarget {
	return el
}

func (el *HTMLElement) AsAttributeTarget() drivers.AttributeTarget {
	return el.attributes
}

func (el *HTMLElement) AsStyleTarget() drivers.StyleTarget {
	return el.styles
}

func (el *HTMLElement) AsValueTarget() drivers.ValueTarget {
	return el
}

func (el *HTMLElement) AsRelationTarget() drivers.RelationTarget {
	return el
}

func (el *HTMLElement) AsInteractionTarget() drivers.InteractionTarget {
	return el
}

func (el *HTMLElement) AsWaitTarget() drivers.WaitTarget {
	return el.wait
}

func (el *HTMLElement) logError(err error) *zerolog.Event {
	return el.logger.
		Error().
		Timestamp().
		Str("objectID", string(el.id)).
		Err(err)
}
