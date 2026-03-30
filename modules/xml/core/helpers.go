package core

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func Root(ctx context.Context, value runtime.Value) (runtime.Value, error) {
	node, nodeType, err := requireXMLNode(ctx, value)
	if err != nil {
		return nil, err
	}

	switch nodeType {
	case nodeTypeDocument:
		return resolveDocumentRoot(ctx, node)
	case nodeTypeElement:
		return value, nil
	case nodeTypeText:
		return runtime.None, nil
	default:
		return nil, invalidXMLNodeType(nodeType)
	}
}

func Text(ctx context.Context, value runtime.Value) (runtime.Value, error) {
	node, nodeType, err := requireXMLNode(ctx, value)
	if err != nil {
		return nil, err
	}

	text, err := collectNodeText(ctx, node, nodeType)
	if err != nil {
		return nil, err
	}

	return runtime.NewString(text), nil
}

func Attr(ctx context.Context, value runtime.Value, name runtime.String) (runtime.Value, error) {
	node, nodeType, err := requireXMLNode(ctx, value)
	if err != nil {
		return nil, err
	}

	switch nodeType {
	case nodeTypeDocument:
		node, err = resolveDocumentRoot(ctx, node)
		if err != nil {
			return nil, err
		}
	case nodeTypeElement:
		// Use the node as-is.
	case nodeTypeText:
		return runtime.None, nil
	default:
		return nil, invalidXMLNodeType(nodeType)
	}

	attrs, err := getOptionalMapField(ctx, node, "attrs")
	if err != nil {
		return nil, err
	}

	attr, err := attrs.Get(ctx, name)
	if err != nil {
		return nil, wrapError(err, fmt.Sprintf("failed to read attribute %q", name.String()))
	}

	if attr == runtime.None {
		return runtime.None, nil
	}

	str, ok := attr.(runtime.String)
	if !ok {
		return nil, newErrorf("attribute %q must be a string", name.String())
	}

	return str, nil
}

func Children(ctx context.Context, value runtime.Value) (runtime.Value, error) {
	node, nodeType, err := requireXMLNode(ctx, value)
	if err != nil {
		return nil, err
	}

	switch nodeType {
	case nodeTypeDocument:
		node, err = resolveDocumentRoot(ctx, node)
		if err != nil {
			return nil, err
		}
	case nodeTypeElement:
		// Use the node as-is.
	case nodeTypeText:
		return runtime.NewArray(0), nil
	default:
		return nil, invalidXMLNodeType(nodeType)
	}

	children, err := getOptionalListField(ctx, node, "children")
	if err != nil {
		return nil, err
	}

	return children, nil
}

func requireXMLNode(ctx context.Context, value runtime.Value) (runtime.Map, string, error) {
	node, ok := value.(runtime.Map)
	if !ok {
		return nil, "", runtime.TypeErrorOf(value, runtime.TypeObject)
	}

	nodeType, err := getRequiredStringField(ctx, node, "type")
	if err != nil {
		return nil, "", err
	}

	switch nodeType {
	case nodeTypeDocument, nodeTypeElement, nodeTypeText:
		return node, nodeType, nil
	default:
		return nil, "", invalidXMLNodeType(nodeType)
	}
}

func resolveDocumentRoot(ctx context.Context, node runtime.Map) (runtime.Map, error) {
	rootValue, err := getRequiredField(ctx, node, "root")
	if err != nil {
		return nil, err
	}

	rootNode, err := asNodeMap(rootValue)
	if err != nil {
		return nil, newError("document root must be an element node")
	}

	rootType, err := getRequiredStringField(ctx, rootNode, "type")
	if err != nil {
		return nil, err
	}

	if rootType != nodeTypeElement {
		return nil, newError("document root must be an element node")
	}

	return rootNode, nil
}

func collectNodeText(ctx context.Context, node runtime.Map, nodeType string) (text string, retErr error) {
	switch nodeType {
	case nodeTypeDocument:
		rootNode, err := resolveDocumentRoot(ctx, node)
		if err != nil {
			return "", err
		}

		return collectNodeText(ctx, rootNode, nodeTypeElement)
	case nodeTypeElement:
		children, err := getOptionalListField(ctx, node, "children")
		if err != nil {
			return "", err
		}

		iter, err := children.Iterate(ctx)
		if err != nil {
			return "", wrapError(err, "failed to iterate XML children")
		}

		defer closeIterator(iter, &retErr)

		var builder strings.Builder

		for {
			child, _, err := iter.Next(ctx)
			if err != nil {
				if err == io.EOF {
					return builder.String(), nil
				}

				return "", wrapError(err, "failed to iterate XML children")
			}

			childNode, childType, err := requireXMLNode(ctx, child)
			if err != nil {
				return "", err
			}

			switch childType {
			case nodeTypeText:
				value, err := getRequiredStringField(ctx, childNode, "value")
				if err != nil {
					return "", err
				}
				builder.WriteString(value)
			case nodeTypeElement:
				text, err := collectNodeText(ctx, childNode, nodeTypeElement)
				if err != nil {
					return "", err
				}
				builder.WriteString(text)
			case nodeTypeDocument:
				return "", newError("document nodes are not allowed inside element children")
			default:
				return "", invalidXMLNodeType(childType)
			}
		}
	case nodeTypeText:
		return getRequiredStringField(ctx, node, "value")
	default:
		return "", invalidXMLNodeType(nodeType)
	}
}

func invalidXMLNodeType(nodeType string) error {
	return newErrorf("expected XML document, element, or text node, got %q", nodeType)
}
