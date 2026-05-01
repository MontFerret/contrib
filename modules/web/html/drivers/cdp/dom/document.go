package dom

import (
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/input"
	"github.com/MontFerret/contrib/modules/web/html/internal/logutil"
)

type HTMLDocument struct {
	logger    zerolog.Logger
	client    *cdp.Client
	dom       *Manager
	input     *input.Manager
	eval      *eval.Runtime
	element   *HTMLElement
	frameTree page.FrameTree
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
	doc.logger = logutil.WithComponent(logger.With(), "html_document").Logger()
	doc.client = client
	doc.dom = domManager
	doc.input = input
	doc.eval = exec
	doc.element = rootElement
	doc.frameTree = frames

	return doc
}

func (doc *HTMLDocument) Close() error {
	return doc.element.Close()
}

func (doc *HTMLDocument) Frame() page.FrameTree {
	return doc.frameTree
}

func (doc *HTMLDocument) Eval() *eval.Runtime {
	return doc.eval
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
