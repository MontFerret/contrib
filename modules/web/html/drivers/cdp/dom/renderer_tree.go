package dom

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mafredri/cdp"
	protocol "github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/page"
	js "github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/rpcc"

	cdpsession "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
)

// Immutable once published. A nil owner denotes the root renderer.
type rendererTree struct {
	tree        page.FrameTree
	owners      map[page.FrameID]*cdpsession.Client
	unavailable map[page.FrameID]error
}

func (v *rendererTree) record(tree page.FrameTree, owner *cdpsession.Client, roots map[page.FrameID]*cdpsession.Client) {
	if renderer := roots[tree.Frame.ID]; renderer != nil {
		owner = renderer
	}

	v.owners[tree.Frame.ID] = owner
	for _, child := range tree.ChildFrames {
		v.record(child, owner, roots)
	}
}

func (v *rendererTree) client(id page.FrameID, root *cdp.Client) *cdp.Client {
	if owner := v.owners[id]; owner != nil {
		return owner.CDP
	}

	return root
}

func (v *rendererTree) preserve(tree page.FrameTree, previous *rendererTree, cause error) {
	v.owners[tree.Frame.ID] = previous.owners[tree.Frame.ID]
	v.unavailable[tree.Frame.ID] = cause
	for _, child := range tree.ChildFrames {
		v.preserve(child, previous, cause)
	}
}

// A renderer omits child processes from its own tree. Verify a missing child's
// owner before retiring a retained handle during the gap between renderers.
func (v *rendererTree) retainMissing(ctx context.Context, node *page.FrameTree, previous *rendererTree, root *cdp.Client, old page.FrameTree) error {
	if old.Frame.ID != node.Frame.ID {
		return nil
	}

	if cause, missing := v.unavailable[node.Frame.ID]; missing {
		*node = old
		v.preserve(old, previous, cause)

		return nil
	}

	for i, child := range old.ChildFrames {
		id := child.Frame.ID
		if _, present := v.owners[id]; present {
			// Chromium usually returns local children in the same order as the previous snapshot.
			if i >= len(node.ChildFrames) || node.ChildFrames[i].Frame.ID != id {
				i = slices.IndexFunc(node.ChildFrames, func(c page.FrameTree) bool { return c.Frame.ID == id })
			}

			if i >= 0 {
				if err := v.retainMissing(ctx, &node.ChildFrames[i], previous, root, child); err != nil {
					return err
				}
			}

			continue
		}

		cause, known := v.unavailable[id]
		if !known {
			_, err := v.client(node.Frame.ID, root).DOM.GetFrameOwner(ctx, protocol.NewGetFrameOwnerArgs(id))
			if err != nil {
				var rpc *rpcc.ResponseError
				if errors.As(err, &rpc) && (strings.Contains(rpc.Message, "not found") || strings.Contains(rpc.Message, "does not belong")) {
					continue
				}

				return err
			}
		}

		v.preserve(child, previous, cause)
		node.ChildFrames = append(node.ChildFrames, child)
	}

	return nil
}

func (v *rendererTree) order(ctx context.Context, node *page.FrameTree, root *cdp.Client) error {
	owner := v.owners[node.Frame.ID]
	mixed := false
	complete := true
	for i := range node.ChildFrames {
		child := &node.ChildFrames[i]
		if _, missing := v.unavailable[child.Frame.ID]; missing {
			complete = false

			continue
		}

		if err := v.order(ctx, child, root); err != nil {
			return err
		}

		mixed = mixed || owner != v.owners[child.Frame.ID]
	}

	if !complete || !mixed || len(node.ChildFrames) < 2 {
		return nil
	}

	ranks := make(map[page.FrameID]int, len(node.ChildFrames))
	for _, child := range node.ChildFrames {
		rank, err := v.frameIndex(ctx, v.client(node.Frame.ID, root), child.Frame.ID)
		if err != nil {
			return err
		}

		ranks[child.Frame.ID] = rank
	}

	sort.SliceStable(node.ChildFrames, func(i, j int) bool { return ranks[node.ChildFrames[i].Frame.ID] < ranks[node.ChildFrames[j].Frame.ID] })

	return nil
}

func (v *rendererTree) frameIndex(ctx context.Context, c *cdp.Client, id page.FrameID) (rank int, err error) {
	owner, err := c.DOM.GetFrameOwner(ctx, protocol.NewGetFrameOwnerArgs(id))
	if err != nil {
		return 0, err
	}

	resolved, err := c.DOM.ResolveNode(ctx, protocol.NewResolveNodeArgs().SetBackendNodeID(owner.BackendNodeID))
	if err != nil {
		return 0, err
	}

	object := resolved.Object.ObjectID
	if object == nil {
		return 0, fmt.Errorf("frame %s has no owner object", id)
	}

	defer func() {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
		defer cancel()
		if releaseErr := c.Runtime.ReleaseObject(cleanup, js.NewReleaseObjectArgs(*object)); releaseErr != nil {
			err = errors.Join(err, releaseErr)
		}
	}()
	const index = `function () {
		const frames = this.ownerDocument.defaultView.frames;
		for (let i = 0; i < frames.length; i++) {
			if (frames[i] === this.contentWindow) return i;
		}
		return -1;
	}`

	result, err := c.Runtime.CallFunctionOn(ctx, js.NewCallFunctionOnArgs(index).SetObjectID(*object).SetReturnByValue(true))
	if err != nil {
		return 0, err
	}

	rank, err = strconv.Atoi(string(result.Result.Value))
	if err != nil || rank < 0 || result.ExceptionDetails != nil {
		return 0, fmt.Errorf("cannot determine sibling index for frame %s", id)
	}

	return rank, nil
}
