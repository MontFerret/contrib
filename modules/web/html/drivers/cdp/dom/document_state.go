package dom

import (
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/input"
)

type (
	documentState struct {
		client  *cdp.Client
		input   *input.Manager
		eval    *eval.Runtime
		element *HTMLElement
		// frameTree is the only mutable generation field and is guarded by HTMLDocument.mu.
		frameTree  page.FrameTree
		generation uint64
	}

	// documentFallback preserves synchronous metadata access after detachment.
	// Lifecycle checks never consult it.
	documentFallback struct {
		element   *HTMLElement
		frameTree page.FrameTree
	}
)
