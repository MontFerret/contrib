package core

import (
	"context"
	"encoding/xml"
	"io"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type elementFrame struct {
	name     string
	node     *runtime.Object
	children *runtime.Array
}

// Decode eagerly decodes XML text into a normalized document object.
func Decode(ctx context.Context, data runtime.String) (runtime.Value, error) {
	decoder := xml.NewDecoder(strings.NewReader(data.String()))

	var (
		root  *runtime.Object
		stack []elementFrame
	)

	for {
		token, err := decoder.RawToken()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, wrapXMLError(err, "failed to decode XML data")
		}

		switch typed := token.(type) {
		case xml.StartElement:
			if len(stack) == 0 && root != nil {
				return nil, newXMLError("multiple root elements are not supported")
			}

			frame := newElementFrame(typed)
			stack = append(stack, frame)
		case xml.EndElement:
			name := xmlNameToString(typed.Name)
			if len(stack) == 0 {
				return nil, newXMLErrorf("unexpected closing tag %q", name)
			}

			frame := stack[len(stack)-1]
			if frame.name != name {
				return nil, newXMLErrorf("mismatched closing tag: expected %q but got %q", frame.name, name)
			}

			stack = stack[:len(stack)-1]

			if len(stack) == 0 {
				root = frame.node
				continue
			}

			if err := stack[len(stack)-1].children.Append(ctx, frame.node); err != nil {
				return nil, err
			}
		case xml.CharData:
			text := string(typed)
			if len(stack) == 0 {
				if strings.TrimSpace(text) == "" {
					continue
				}

				return nil, newXMLError("text outside root element is not supported")
			}

			if err := stack[len(stack)-1].children.Append(ctx, newTextNode(text)); err != nil {
				return nil, err
			}
		case xml.Comment, xml.Directive, xml.ProcInst:
			continue
		}
	}

	if len(stack) > 0 {
		return nil, newXMLErrorf("unexpected EOF while closing %q", stack[len(stack)-1].name)
	}

	if root == nil {
		return nil, newXMLError("document has no root element")
	}

	return newDocumentNode(root), nil
}

func newElementFrame(token xml.StartElement) elementFrame {
	name := xmlNameToString(token.Name)
	attrs := newAttrsObject(token.Attr)
	children := runtime.NewArray(0)

	return elementFrame{
		name:     name,
		node:     newElementNode(name, attrs, children),
		children: children,
	}
}
