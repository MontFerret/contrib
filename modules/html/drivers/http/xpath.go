package http

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"golang.org/x/net/html"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func EvalXPathToNode(selection *goquery.Selection, expression string) (drivers.HTMLNode, error) {
	node := htmlquery.FindOne(fromSelectionToNode(selection), expression)

	if node == nil {
		return nil, nil
	}

	return parseXPathNode(node)
}

func EvalXPathToElement(selection *goquery.Selection, expression string) (drivers.HTMLElement, error) {
	node, err := EvalXPathToNode(selection, expression)

	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, nil
	}

	return drivers.ToElement(node)
}

func EvalXPathToNodes(selection *goquery.Selection, expression string) (runtime.List, error) {
	return EvalXPathToNodesWith(selection, expression, func(node *html.Node) (runtime.Value, error) {
		return parseXPathNode(node)
	})
}

func EvalXPathToNodesWith(selection *goquery.Selection, expression string, mapper func(node *html.Node) (runtime.Value, error)) (runtime.List, error) {
	out, err := evalXPathToInternal(selection, expression)

	if err != nil {
		return nil, err
	}

	switch res := out.(type) {
	case *xpath.NodeIterator:
		items := runtime.NewArray(10)
		ctx := context.Background()

		for res.MoveNext() {
			item, err := mapper(res.Current().(*htmlquery.NodeNavigator).Current())

			if err != nil {
				return nil, err
			}

			if item != nil {
				_ = items.Append(ctx, item)
			}
		}

		return items, nil
	default:
		return runtime.EmptyArray(), nil
	}
}

func EvalXPathTo(selection *goquery.Selection, expression string) (runtime.Value, error) {
	out, err := evalXPathToInternal(selection, expression)

	if err != nil {
		return nil, err
	}

	switch res := out.(type) {
	case *xpath.NodeIterator:
		items := runtime.NewArray(10)
		ctx := context.Background()

		for res.MoveNext() {
			var item runtime.Value

			node := res.Current()

			switch node.NodeType() {
			case xpath.TextNode:
				item = runtime.NewString(node.Value())
			case xpath.AttributeNode:
				item = runtime.NewString(node.Value())
			default:
				i, err := parseXPathNode(node.(*htmlquery.NodeNavigator).Current())

				if err != nil {
					return nil, err
				}

				item = i
			}

			if item != nil {
				_ = items.Append(ctx, item)
			}
		}

		return items, nil
	default:
		return runtime.Parse(res), nil
	}
}

func evalXPathToInternal(selection *goquery.Selection, expression string) (any, error) {
	exp, err := xpath.Compile(expression)

	if err != nil {
		return nil, err
	}

	return exp.Evaluate(htmlquery.CreateXPathNavigator(fromSelectionToNode(selection))), nil
}

func parseXPathNode(node *html.Node) (drivers.HTMLNode, error) {
	if node == nil {
		return nil, nil
	}

	switch node.Type {
	case html.DocumentNode:
		url := htmlquery.SelectAttr(node, "url")
		return NewHTMLDocument(goquery.NewDocumentFromNode(node), url, nil)
	case html.ElementNode:
		return NewHTMLElement(&goquery.Selection{Nodes: []*html.Node{node}})
	default:
		return nil, nil
	}
}
