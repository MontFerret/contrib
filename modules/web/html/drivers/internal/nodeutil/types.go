package nodeutil

import (
	"golang.org/x/net/html"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func FromHTMLType(nodeType html.NodeType) int {
	switch nodeType {
	case html.DocumentNode:
		return 9
	case html.ElementNode:
		return 1
	case html.TextNode:
		return 3
	case html.CommentNode:
		return 8
	case html.DoctypeNode:
		return 10
	default:
		return 0
	}
}

func IsEmptyValue(value runtime.Value) bool {
	if value == nil {
		return true
	}

	return value == runtime.None
}
