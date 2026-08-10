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
	styles     *elementStyles
	dataset    *elementDataset
	input      *input.Manager
	eval       *elementExecutor
	attributes *elementAttributes
	client     *cdp.Client
	classes    *elementClasses
	dom        *Manager
	wait       *elementWait
	nodeType   *lazy.Value
	nodeName   *lazy.Value
	document   *HTMLDocument
	id         cdpruntime.RemoteObjectID
	generation uint64
}

func NewHTMLElement(
	logger zerolog.Logger,
	client *cdp.Client,
	domManager *Manager,
	input *input.Manager,
	exec *eval.Runtime,
	id cdpruntime.RemoteObjectID,
) *HTMLElement {
	return newHTMLElement(logger, client, domManager, input, exec, id, nil, 0)
}

func newHTMLElement(
	logger zerolog.Logger,
	client *cdp.Client,
	domManager *Manager,
	input *input.Manager,
	exec *eval.Runtime,
	id cdpruntime.RemoteObjectID,
	document *HTMLDocument,
	generation uint64,
) *HTMLElement {
	el := new(HTMLElement)
	el.logger = logutil.WithComponent(logger.With(), "dom_element").
		Str("object_id", string(id)).
		Logger()
	el.client = client
	el.dom = domManager
	el.input = input
	el.eval = newElementExecutor(exec, document, generation)
	el.id = id
	el.document = document
	el.generation = generation
	el.attributes = newElementAttributes(el.eval, id)
	el.styles = newElementStyles(el.eval, id)
	el.classes = newElementClasses(el.eval, id)
	el.dataset = newElementDataset(el.eval, id)
	el.wait = newElementWait(el.eval, id)
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

func (el *HTMLElement) ensureAttached() error {
	if el.eval == nil {
		return drivers.ErrDetached
	}

	return el.eval.ensureAttached()
}

func (el *HTMLElement) normalizeError(ctx context.Context, err error) error {
	if el.eval == nil {
		return drivers.ErrDetached
	}

	return el.eval.normalizeError(ctx, err)
}

func (el *HTMLElement) bindDocument(document *HTMLDocument, exec *eval.Runtime, generation uint64) {
	el.document = document
	el.generation = generation
	el.eval = newElementExecutor(exec, document, generation)
	el.attributes = newElementAttributes(el.eval, el.id)
	el.styles = newElementStyles(el.eval, el.id)
	el.classes = newElementClasses(el.eval, el.id)
	el.dataset = newElementDataset(el.eval, el.id)
	el.wait = newElementWait(el.eval, el.id)
}

func (el *HTMLElement) logError(err error) *zerolog.Event {
	return el.logger.
		Error().
		Timestamp().
		Str("objectID", string(el.id)).
		Err(err)
}
