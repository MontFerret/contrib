package core

import (
	"encoding/xml"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const (
	nodeTypeDocument      = "document"
	nodeTypeElement       = "element"
	nodeTypeText          = "text"
	eventTypeStartElement = "startElement"
	eventTypeEndElement   = "endElement"
)

func newDocumentNode(root *runtime.Object) *runtime.Object {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"type": runtime.NewString(nodeTypeDocument),
		"root": root,
	})
}

func newElementNode(name string, attrs *runtime.Object, children *runtime.Array) *runtime.Object {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"type":     runtime.NewString(nodeTypeElement),
		"name":     runtime.NewString(name),
		"attrs":    attrs,
		"children": children,
	})
}

func newTextNode(value string) *runtime.Object {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"type":  runtime.NewString(nodeTypeText),
		"value": runtime.NewString(value),
	})
}

func newStartElementEvent(name string, attrs *runtime.Object) *runtime.Object {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"type":  runtime.NewString(eventTypeStartElement),
		"name":  runtime.NewString(name),
		"attrs": attrs,
	})
}

func newEndElementEvent(name string) *runtime.Object {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"type": runtime.NewString(eventTypeEndElement),
		"name": runtime.NewString(name),
	})
}

func newAttrsObject(attrs []xml.Attr) *runtime.Object {
	props := make(map[string]runtime.Value, len(attrs))

	for _, attr := range attrs {
		props[xmlNameToString(attr.Name)] = runtime.NewString(attr.Value)
	}

	return runtime.NewObjectWith(props)
}
