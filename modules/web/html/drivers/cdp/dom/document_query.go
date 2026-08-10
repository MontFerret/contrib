package dom

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (doc *HTMLDocument) GetChildNodes(ctx context.Context) (runtime.List, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.List, error) {
		return state.element.GetChildNodes(ctx)
	})
}

func (doc *HTMLDocument) GetChildNode(ctx context.Context, idx runtime.Int) (runtime.Value, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.Value, error) {
		return state.element.GetChildNode(ctx, idx)
	})
}

func (doc *HTMLDocument) QuerySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.Value, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.Value, error) {
		return state.element.QuerySelector(ctx, selector)
	})
}

func (doc *HTMLDocument) QuerySelectorAll(ctx context.Context, selector drivers.QuerySelector) (runtime.List, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.List, error) {
		return state.element.QuerySelectorAll(ctx, selector)
	})
}

func (doc *HTMLDocument) CountBySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.Int, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.Int, error) {
		return state.element.CountBySelector(ctx, selector)
	})
}

func (doc *HTMLDocument) ExistsBySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.Boolean, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.Boolean, error) {
		return state.element.ExistsBySelector(ctx, selector)
	})
}

func (doc *HTMLDocument) GetParentDocument(ctx context.Context) (drivers.HTMLDocument, error) {
	frameTree, err := doc.snapshotFrame()
	if err != nil {
		return nil, err
	}

	if frameTree.Frame.ParentID == nil {
		return nil, nil
	}

	return doc.dom.GetFrameNode(ctx, *frameTree.Frame.ParentID)
}

func (doc *HTMLDocument) GetChildDocuments(ctx context.Context) (runtime.List, error) {
	frameTree, err := doc.snapshotFrame()
	if err != nil {
		return nil, err
	}

	arr := runtime.NewArray(len(frameTree.ChildFrames))

	for _, childFrame := range frameTree.ChildFrames {
		frame, err := doc.dom.GetFrameNode(ctx, childFrame.Frame.ID)
		if err != nil {
			return nil, err
		}

		if frame != nil {
			_ = arr.Append(ctx, frame)
		}
	}

	return arr, nil
}

func (doc *HTMLDocument) XPath(ctx context.Context, expression runtime.String) (runtime.Value, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.Value, error) {
		return state.element.XPath(ctx, expression)
	})
}

func (doc *HTMLDocument) Query(ctx context.Context, q runtime.Query) (runtime.List, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.List, error) {
		return state.element.Query(ctx, q)
	})
}

func (doc *HTMLDocument) QueryOne(ctx context.Context, q runtime.Query) (runtime.Value, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.Value, error) {
		return state.element.QueryOne(ctx, q)
	})
}

func (doc *HTMLDocument) QueryCount(ctx context.Context, q runtime.Query) (runtime.Int, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.Int, error) {
		return state.element.QueryCount(ctx, q)
	})
}

func (doc *HTMLDocument) QueryExists(ctx context.Context, q runtime.Query) (runtime.Boolean, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.Boolean, error) {
		return state.element.QueryExists(ctx, q)
	})
}
