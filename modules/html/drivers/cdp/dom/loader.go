package dom

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers/cdp/eval"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/mafredri/cdp/protocol/page"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
)

type NodeLoader struct {
	dom *Manager
}

func NewNodeLoader(dom *Manager) eval.ValueLoader {
	return &NodeLoader{dom}
}

func (n *NodeLoader) Load(ctx context.Context, frameID page.FrameID, _ eval.RemoteObjectType, _ eval.RemoteClassName, id cdpruntime.RemoteObjectID) (runtime.Value, error) {
	return n.dom.ResolveElement(ctx, frameID, id)
}
