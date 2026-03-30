package core

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"sort"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Encode serializes a normalized XML document or element into XML text.
func Encode(ctx context.Context, value runtime.Value) (string, error) {
	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)

	if err := encodeTopLevel(ctx, encoder, value); err != nil {
		return "", err
	}

	if err := encoder.Flush(); err != nil {
		return "", wrapError(err, "failed to flush XML output")
	}

	return buf.String(), nil
}

func encodeTopLevel(ctx context.Context, encoder *xml.Encoder, value runtime.Value) error {
	node, err := asNodeMap(value)
	if err != nil {
		return newError("encode expects a document or element node")
	}

	nodeType, err := getRequiredStringField(ctx, node, "type")
	if err != nil {
		return err
	}

	switch nodeType {
	case nodeTypeDocument:
		rootValue, err := getRequiredField(ctx, node, "root")
		if err != nil {
			return err
		}

		rootNode, err := asNodeMap(rootValue)
		if err != nil {
			return newError("document root must be an element node")
		}

		rootType, err := getRequiredStringField(ctx, rootNode, "type")
		if err != nil {
			return err
		}

		if rootType != nodeTypeElement {
			return newError("document root must be an element node")
		}

		return encodeElement(ctx, encoder, rootNode)
	case nodeTypeElement:
		return encodeElement(ctx, encoder, node)
	default:
		return newError("encode expects a document or element node")
	}
}

func encodeNode(ctx context.Context, encoder *xml.Encoder, value runtime.Value) error {
	node, err := asNodeMap(value)
	if err != nil {
		return newError("element children must be element or text nodes")
	}

	nodeType, err := getRequiredStringField(ctx, node, "type")
	if err != nil {
		return err
	}

	switch nodeType {
	case nodeTypeElement:
		return encodeElement(ctx, encoder, node)
	case nodeTypeText:
		return encodeText(ctx, encoder, node)
	case nodeTypeDocument:
		return newError("document nodes are not allowed inside element children")
	default:
		return newErrorf("unsupported XML node type %q", nodeType)
	}
}

func encodeElement(ctx context.Context, encoder *xml.Encoder, node runtime.Map) error {
	name, err := getRequiredStringField(ctx, node, "name")
	if err != nil {
		return err
	}

	xmlName, err := xmlNameFromString(name)
	if err != nil {
		return err
	}

	attrsValue, err := getOptionalMapField(ctx, node, "attrs")
	if err != nil {
		return err
	}

	attrs, err := collectAttrs(ctx, attrsValue)
	if err != nil {
		return err
	}

	start := xml.StartElement{
		Name: xmlName,
		Attr: attrs,
	}

	if err := encoder.EncodeToken(start); err != nil {
		return wrapError(err, fmt.Sprintf("failed to encode start element %q", name))
	}

	children, err := getOptionalListField(ctx, node, "children")
	if err != nil {
		return err
	}

	if err := encodeChildren(ctx, encoder, children); err != nil {
		return err
	}

	if err := encoder.EncodeToken(start.End()); err != nil {
		return wrapError(err, fmt.Sprintf("failed to encode end element %q", name))
	}

	return nil
}

func encodeText(ctx context.Context, encoder *xml.Encoder, node runtime.Map) error {
	value, err := getRequiredStringField(ctx, node, "value")
	if err != nil {
		return err
	}

	if err := encoder.EncodeToken(xml.CharData([]byte(value))); err != nil {
		return wrapError(err, "failed to encode text node")
	}

	return nil
}

func encodeChildren(ctx context.Context, encoder *xml.Encoder, children runtime.List) (retErr error) {
	iter, err := children.Iterate(ctx)
	if err != nil {
		return wrapError(err, "failed to iterate XML children")
	}

	defer closeIterator(iter, &retErr)

	for {
		child, _, err := iter.Next(ctx)
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return wrapError(err, "failed to iterate XML children")
		}

		if err := encodeNode(ctx, encoder, child); err != nil {
			return err
		}
	}
}

func collectAttrs(ctx context.Context, attrs runtime.Map) ([]xml.Attr, error) {
	type pair struct {
		name  string
		value string
	}

	pairs := make([]pair, 0)

	if err := attrs.ForEach(ctx, func(_ context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		attrValue, ok := value.(runtime.String)
		if !ok {
			return false, newErrorf("attribute %q must be a string", key.String())
		}

		pairs = append(pairs, pair{
			name:  key.String(),
			value: attrValue.String(),
		})

		return true, nil
	}); err != nil {
		return nil, err
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].name < pairs[j].name
	})

	out := make([]xml.Attr, 0, len(pairs))

	for _, pair := range pairs {
		name, err := xmlNameFromString(pair.name)
		if err != nil {
			return nil, err
		}

		out = append(out, xml.Attr{
			Name:  name,
			Value: pair.value,
		})
	}

	return out, nil
}

func getRequiredField(ctx context.Context, node runtime.Map, key string) (runtime.Value, error) {
	value, err := node.Get(ctx, runtime.NewString(key))
	if err != nil {
		return nil, wrapError(err, fmt.Sprintf("failed to read field %q", key))
	}

	if value == runtime.None {
		return nil, newErrorf("field %q is required", key)
	}

	return value, nil
}

func getRequiredStringField(ctx context.Context, node runtime.Map, key string) (string, error) {
	value, err := getRequiredField(ctx, node, key)
	if err != nil {
		return "", err
	}

	str, ok := value.(runtime.String)
	if !ok {
		return "", newErrorf("field %q must be a string", key)
	}

	return str.String(), nil
}

func getOptionalMapField(ctx context.Context, node runtime.Map, key string) (runtime.Map, error) {
	value, err := node.Get(ctx, runtime.NewString(key))
	if err != nil {
		return nil, wrapError(err, fmt.Sprintf("failed to read field %q", key))
	}

	if value == runtime.None {
		return runtime.NewObject(), nil
	}

	attrs, ok := value.(runtime.Map)
	if !ok {
		return nil, newErrorf("field %q must be an object", key)
	}

	return attrs, nil
}

func getOptionalListField(ctx context.Context, node runtime.Map, key string) (runtime.List, error) {
	value, err := node.Get(ctx, runtime.NewString(key))
	if err != nil {
		return nil, wrapError(err, fmt.Sprintf("failed to read field %q", key))
	}

	if value == runtime.None {
		return runtime.NewArray(0), nil
	}

	list, ok := value.(runtime.List)
	if !ok {
		return nil, newErrorf("field %q must be an array", key)
	}

	return list, nil
}

func asNodeMap(value runtime.Value) (runtime.Map, error) {
	node, ok := value.(runtime.Map)
	if !ok {
		return nil, newError("expected XML node object")
	}

	return node, nil
}

func closeIterator(iter runtime.Iterator, retErr *error) {
	closer, ok := iter.(io.Closer)
	if !ok {
		return
	}

	if err := closer.Close(); err != nil && *retErr == nil {
		*retErr = wrapError(err, "failed to close XML iterator")
	}
}
