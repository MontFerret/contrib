package core

import (
	"context"
	"io"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type elementFrame struct {
	node     *runtime.Object
	children *runtime.Array
	name     string
}

// Decode eagerly decodes XML text into a normalized document object.
func Decode(ctx context.Context, data runtime.String) (runtime.Value, error) {
	cursor := newDecodeCursorFromReader(strings.NewReader(data.String()))

	var (
		root  *runtime.Object
		stack []elementFrame
	)

	for {
		event, err := cursor.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		switch event.kind {
		case decodeEventStart:
			frame := newElementFrame(event.name, event.attrs)
			stack = append(stack, frame)
		case decodeEventEnd:
			frame := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if len(stack) == 0 {
				root = frame.node
				continue
			}

			if err := stack[len(stack)-1].children.Append(ctx, frame.node); err != nil {
				return nil, err
			}
		case decodeEventText:
			if err := stack[len(stack)-1].children.Append(ctx, newTextNode(event.text)); err != nil {
				return nil, err
			}
		}
	}

	if root == nil {
		return nil, newError("document has no root element")
	}

	return newDocumentNode(root), nil
}

func newElementFrame(name string, attrs *runtime.Object) elementFrame {
	children := runtime.NewArray(0)

	return elementFrame{
		name:     name,
		node:     newElementNode(name, attrs, children),
		children: children,
	}
}
