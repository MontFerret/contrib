package dom

import (
	"context"
	"errors"
	"time"

	"github.com/mafredri/cdp/protocol/page"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	cdpsession "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
)

func (m *Manager) frameTree(ctx context.Context, strict bool, required page.FrameID) (*page.GetFrameTreeReply, error) {
	if !m.sessions.IsFlat() {
		return m.rootClient.Page.GetFrameTree(ctx)
	}

	ctx, cancel := context.WithTimeout(ctx, drivers.DefaultTimeout*time.Millisecond)
	defer cancel()
	for {
		previous := m.renderers.Load()

		view, err := m.readRendererTree(ctx, previous)
		if err != nil {
			return nil, err
		}

		if !m.renderers.CompareAndSwap(previous, view) {
			continue
		}

		problem, waiting := view.unavailable[required]
		if strict {
			var failures []error
			for _, cause := range view.unavailable {
				if cause != nil {
					failures = append(failures, cause)
				}
			}

			problem = errors.Join(failures...)
			waiting = len(view.unavailable) > 0
		}

		if problem != nil {
			return nil, problem
		}

		if !waiting {
			return &page.GetFrameTreeReply{FrameTree: view.tree}, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-m.sessions.PageContext().Done():
			return nil, context.Cause(m.sessions.PageContext())
		case <-time.After(20 * time.Millisecond):
		}
	}
}

func (m *Manager) readRendererTree(ctx context.Context, previous *rendererTree) (*rendererTree, error) {
	frames, transitions := m.sessions.FrameTopology()

	root, err := m.rootClient.Page.GetFrameTree(ctx)
	if err != nil {
		return nil, err
	}

	v := &rendererTree{tree: root.FrameTree, owners: make(map[page.FrameID]*cdpsession.Client), unavailable: make(map[page.FrameID]error)}
	roots := make(map[page.FrameID]*cdpsession.Client, len(frames))
	var remote []page.FrameTree
	for _, c := range frames {
		reply, err := c.CDP.Page.GetFrameTree(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			// The session can disappear between the topology snapshot and this reply.
			// Only a published initialization failure is terminal for the document.
			v.unavailable[page.FrameID(c.TargetID)] = nil

			continue
		}

		roots[reply.FrameTree.Frame.ID] = c
		remote = append(remote, reply.FrameTree)
	}

	for id, cause := range transitions {
		v.unavailable[page.FrameID(id)] = cause
	}

	if len(remote) > 0 {
		mergeRemoteFrames(&v.tree, remote)
	}

	v.record(v.tree, nil, roots)
	if previous != nil {
		if err := v.retainMissing(ctx, &v.tree, previous, m.rootClient, previous.tree); err != nil {
			return nil, err
		}
	}

	if len(remote) > 0 {
		if err := v.order(ctx, &v.tree, m.rootClient); err != nil {
			return nil, err
		}
	}

	return v, nil
}
