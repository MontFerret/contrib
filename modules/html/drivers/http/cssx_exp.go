package http

import (
	"bytes"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MontFerret/contrib/modules/html/drivers/common/cssx"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

var cssxNonNumeric = regexp.MustCompile(`[^\d+\-.,eE]`)

func cssxQueryAll(selection *goquery.Selection, selector string) []any {
	if selection == nil {
		return []any{}
	}

	return cssxNodesToAny(selection.Find(selector).Nodes)
}

func cssxApplyCall(name cssx.Expression, args []any, values []any, baseURL *url.URL) any {
	var input any

	if len(values) > 0 {
		input = values[len(values)-1]
	}

	switch name {
	case cssx.ExpressionFirst:
		arr := cssxToArray(input)

		if len(arr) == 0 {
			return nil
		}

		return arr[0]
	case cssx.ExpressionLast:
		arr := cssxToArray(input)

		if len(arr) == 0 {
			return nil
		}

		return arr[len(arr)-1]
	case cssx.ExpressionNth:
		arr := cssxToArray(input)
		idx, ok := cssxToInt(cssxArgNumber(args, 0))

		if !ok || idx < 0 || idx >= len(arr) {
			return nil
		}

		return arr[idx]
	case cssx.ExpressionTake:
		arr := cssxToArray(input)
		n, ok := cssxToInt(cssxArgNumber(args, 0))

		if !ok || n <= 0 {
			return []any{}
		}

		if n > len(arr) {
			n = len(arr)
		}

		out := make([]any, n)
		copy(out, arr[:n])

		return out
	case cssx.ExpressionSkip:
		arr := cssxToArray(input)
		n, ok := cssxToInt(cssxArgNumber(args, 0))

		if !ok || n <= 0 {
			return arr
		}

		if n >= len(arr) {
			return []any{}
		}

		out := make([]any, len(arr)-n)
		copy(out, arr[n:])

		return out
	case cssx.ExpressionSlice:
		arr := cssxToArray(input)
		start, startOk := cssxToInt(cssxArgNumber(args, 0))
		count, countOk := cssxToInt(cssxArgNumber(args, 1))

		if !startOk || !countOk || count <= 0 {
			return []any{}
		}

		return cssxSliceArray(arr, start, start+count)
	case cssx.ExpressionWithin:
		var scope any

		if len(values) > 1 {
			scope = values[0]
		}

		scopedNodes := cssxToNodes(scope)

		if len(scopedNodes) == 0 {
			return []any{}
		}

		if selector, ok := input.(string); ok {
			var out []*html.Node

			for _, node := range scopedNodes {
				doc := goquery.NewDocumentFromNode(node)
				out = append(out, doc.Selection.Find(selector).Nodes...)
			}

			return cssxNodesToAny(cssxDedupeNodes(out))
		}

		nodeValues := cssxToNodes(input)

		if len(nodeValues) > 0 {
			out := make([]any, 0, len(nodeValues))

			for _, node := range nodeValues {
				for _, scopeNode := range scopedNodes {
					if cssxContainsNode(scopeNode, node) {
						out = append(out, node)
						break
					}
				}
			}

			return out
		}

		return input
	case cssx.ExpressionParent:
		node := cssxFirstNode(input)

		if node == nil {
			return nil
		}

		return cssxParentElement(node)
	case cssx.ExpressionClosest:
		node := cssxFirstNode(input)

		if node == nil {
			return nil
		}

		var candidates []*html.Node

		if len(values) > 1 {
			candidates = cssxToNodes(values[0])
		}

		if len(candidates) == 0 {
			return cssxParentElement(node)
		}

		candidateSet := make(map[*html.Node]struct{}, len(candidates))

		for _, candidate := range candidates {
			candidateSet[candidate] = struct{}{}
		}

		for cursor := node; cursor != nil; cursor = cssxParentElement(cursor) {
			if _, ok := candidateSet[cursor]; ok {
				return cursor
			}
		}

		return nil
	case cssx.ExpressionChildren:
		node := cssxFirstNode(input)

		if node == nil {
			return []any{}
		}

		children := cssxElementChildren(node)

		if len(values) <= 1 {
			return cssxNodesToAny(children)
		}

		candidates := cssxToNodes(values[0])

		if len(candidates) == 0 {
			return cssxNodesToAny(children)
		}

		candidateSet := make(map[*html.Node]struct{}, len(candidates))

		for _, candidate := range candidates {
			candidateSet[candidate] = struct{}{}
		}

		out := make([]any, 0, len(children))

		for _, child := range children {
			if _, ok := candidateSet[child]; ok {
				out = append(out, child)
			}
		}

		return out
	case cssx.ExpressionNext:
		node := cssxFirstNode(input)

		if node == nil {
			return nil
		}

		next := cssxNextElementSibling(node)

		if next == nil {
			return nil
		}

		if len(values) <= 1 {
			return next
		}

		candidates := cssxToNodes(values[0])

		if len(candidates) == 0 {
			return next
		}

		for _, candidate := range candidates {
			if candidate == next {
				return next
			}
		}

		return nil
	case cssx.ExpressionPrev:
		node := cssxFirstNode(input)

		if node == nil {
			return nil
		}

		prev := cssxPrevElementSibling(node)

		if prev == nil {
			return nil
		}

		if len(values) <= 1 {
			return prev
		}

		candidates := cssxToNodes(values[0])

		if len(candidates) == 0 {
			return prev
		}

		for _, candidate := range candidates {
			if candidate == prev {
				return prev
			}
		}

		return nil
	case cssx.ExpressionExists:
		return len(cssxToArray(input)) > 0
	case cssx.ExpressionEmpty:
		return len(cssxToArray(input)) == 0
	case cssx.ExpressionHas:
		node := cssxFirstNode(input)

		if node == nil {
			return false
		}

		var candidates []*html.Node

		if len(values) > 1 {
			candidates = cssxToNodes(values[0])
		}

		for _, candidate := range candidates {
			if cssxContainsNode(node, candidate) {
				return true
			}
		}

		return false
	case cssx.ExpressionMatches:
		node := cssxFirstNode(input)

		if node == nil {
			return false
		}

		var candidates []*html.Node

		if len(values) > 1 {
			candidates = cssxToNodes(values[0])
		}

		for _, candidate := range candidates {
			if candidate == node {
				return true
			}
		}

		return false
	case cssx.ExpressionCount:
		return len(cssxToArray(input))
	case cssx.ExpressionIndexOf:
		var list []any

		if len(values) > 1 {
			list = cssxToArray(values[0])
		}

		item := any(cssxFirstNode(input))

		if item == nil {
			item = input
		}

		for idx, candidate := range list {
			if cssxAnyEqual(candidate, item) {
				return idx
			}
		}

		return -1
	case cssx.ExpressionLen:
		switch v := input.(type) {
		case string:
			return len(v)
		case []any:
			return len(v)
		case nil:
			return 0
		default:
			return len(cssxToArray(v))
		}
	case cssx.ExpressionText:
		node := cssxFirstNode(input)

		if node == nil {
			return nil
		}

		return cssxTextContent(node)
	case cssx.ExpressionTexts:
		nodes := cssxToNodes(input)
		out := make([]any, 0, len(nodes))

		for _, node := range nodes {
			out = append(out, cssxTextContent(node))
		}

		return out
	case cssx.ExpressionOwnText:
		node := cssxFirstNode(input)

		if node == nil {
			return nil
		}

		return cssxOwnText(node)
	case cssx.ExpressionNormalize:
		return cssxNormalizeSpace(cssxTextOf(input))
	case cssx.ExpressionTrim:
		return strings.TrimSpace(cssxTextOf(input))
	case cssx.ExpressionJoin:
		sep := cssxArgString(args, 0)
		arr := cssxToArray(input)
		vals := make([]string, 0, len(arr))

		for _, item := range arr {
			vals = append(vals, cssxTextOf(item))
		}

		return strings.Join(vals, sep)
	case cssx.ExpressionAttr:
		node := cssxFirstNode(input)

		if node == nil {
			return nil
		}

		if val, ok := cssxNodeAttr(node, cssxArgString(args, 0)); ok {
			return val
		}

		return nil
	case cssx.ExpressionAttrs:
		nodes := cssxToNodes(input)
		name := cssxArgString(args, 0)
		out := make([]any, 0, len(nodes))

		for _, node := range nodes {
			if val, ok := cssxNodeAttr(node, name); ok {
				out = append(out, val)
				continue
			}

			out = append(out, nil)
		}

		return out
	case cssx.ExpressionProp:
		node := cssxFirstNode(input)

		if node == nil {
			return nil
		}

		return cssxNodeProp(node, cssxArgString(args, 0))
	case cssx.ExpressionHTML:
		node := cssxFirstNode(input)

		if node == nil {
			return nil
		}

		return cssxInnerHTML(node)
	case cssx.ExpressionOuterHTML:
		node := cssxFirstNode(input)

		if node == nil {
			return nil
		}

		return cssxOuterHTML(node)
	case cssx.ExpressionValue:
		node := cssxFirstNode(input)

		if node == nil {
			return nil
		}

		return cssxNodeProp(node, "value")
	case cssx.ExpressionAbsURL:
		arr := cssxToArray(input)

		if len(arr) == 0 {
			return nil
		}

		return cssxAsURL(arr[0], baseURL)
	case cssx.ExpressionURL:
		node := cssxFirstNode(input)

		if node == nil {
			return nil
		}

		if val, ok := cssxNodeAttr(node, cssxArgString(args, 0)); ok {
			return cssxAsURL(val, baseURL)
		}

		return nil
	case cssx.ExpressionParseURL:
		arr := cssxToArray(input)

		if len(arr) == 0 {
			return nil
		}

		return cssxParseURL(arr[0], baseURL)
	case cssx.ExpressionFilter:
		predicate := any(true)
		source := input

		if len(values) > 1 {
			predicate = values[0]
			source = values[1]
		}

		arr := cssxToArray(source)

		switch v := predicate.(type) {
		case []any:
			out := make([]any, 0, len(arr))

			for _, item := range arr {
				if cssxArrayContains(v, item) {
					out = append(out, item)
				}
			}

			return out
		case bool:
			if v {
				return arr
			}

			return []any{}
		case string:
			out := make([]any, 0, len(arr))

			for _, item := range arr {
				if strings.Contains(cssxTextOf(item), v) {
					out = append(out, item)
				}
			}

			return out
		case nil:
			return []any{}
		default:
			out := make([]any, 0, len(arr))

			for _, item := range arr {
				if cssxTruthy(item) {
					out = append(out, item)
				}
			}

			return out
		}
	case cssx.ExpressionWithAttr:
		nodes := cssxToNodes(input)
		name := cssxArgString(args, 0)
		out := make([]any, 0, len(nodes))

		for _, node := range nodes {
			if cssxNodeHasAttr(node, name) {
				out = append(out, node)
			}
		}

		return out
	case cssx.ExpressionWithText:
		nodes := cssxToNodes(input)
		needle := cssxArgString(args, 0)
		out := make([]any, 0, len(nodes))

		for _, node := range nodes {
			if strings.Contains(cssxTextContent(node), needle) {
				out = append(out, node)
			}
		}

		return out
	case cssx.ExpressionDedupeByAttr:
		nodes := cssxToNodes(input)
		name := cssxArgString(args, 0)
		seen := make(map[string]struct{}, len(nodes))
		out := make([]any, 0, len(nodes))

		for _, node := range nodes {
			val, ok := cssxNodeAttr(node, name)

			key := "<nil>"
			if ok {
				key = val
			}

			if _, exists := seen[key]; exists {
				continue
			}

			seen[key] = struct{}{}
			out = append(out, node)
		}

		return out
	case cssx.ExpressionDedupeByText:
		nodes := cssxToNodes(input)
		seen := make(map[string]struct{}, len(nodes))
		out := make([]any, 0, len(nodes))

		for _, node := range nodes {
			key := cssxNormalizeSpace(cssxTextContent(node))

			if _, exists := seen[key]; exists {
				continue
			}

			seen[key] = struct{}{}
			out = append(out, node)
		}

		return out
	case cssx.ExpressionReplace:
		pattern := cssxArgString(args, 0)
		replacement := cssxArgString(args, 1)
		source := cssxTextOf(input)
		rx, err := regexp.Compile(pattern)

		if err == nil {
			return rx.ReplaceAllString(source, replacement)
		}

		return strings.ReplaceAll(source, pattern, replacement)
	case cssx.ExpressionRegex:
		pattern := cssxArgString(args, 0)
		group := 0

		if len(args) > 1 {
			if value, ok := cssxToInt(cssxArgNumber(args, 1)); ok {
				group = value
			}
		}

		source := cssxTextOf(input)
		rx, err := regexp.Compile(pattern)

		if err != nil {
			return nil
		}

		matches := rx.FindStringSubmatch(source)

		if len(matches) == 0 || group < 0 || group >= len(matches) {
			return nil
		}

		return matches[group]
	case cssx.ExpressionToNumber:
		return cssxToNumber(input)
	case cssx.ExpressionToDate:
		return cssxToDate(input, args)
	default:
		return []any{}
	}
}

func cssxArgString(args []any, idx int) string {
	if idx < 0 || idx >= len(args) {
		return ""
	}

	if str, ok := args[idx].(string); ok {
		return str
	}

	return fmt.Sprint(args[idx])
}

func cssxArgNumber(args []any, idx int) float64 {
	if idx < 0 || idx >= len(args) {
		return math.NaN()
	}

	switch v := args[idx].(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return math.NaN()
	}
}

func cssxToInt(value float64) (int, bool) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, false
	}

	rounded := math.Trunc(value)

	if rounded != value {
		return 0, false
	}

	return int(rounded), true
}

func cssxToArray(value any) []any {
	switch v := value.(type) {
	case nil:
		return []any{}
	case []any:
		out := make([]any, 0, len(v))

		for _, item := range v {
			if item != nil {
				out = append(out, item)
			}
		}

		return out
	default:
		return []any{v}
	}
}

func cssxToNodes(value any) []*html.Node {
	arr := cssxToArray(value)
	out := make([]*html.Node, 0, len(arr))

	for _, item := range arr {
		node, ok := item.(*html.Node)

		if ok && node != nil {
			out = append(out, node)
		}
	}

	return out
}

func cssxFirstNode(value any) *html.Node {
	if node, ok := value.(*html.Node); ok {
		return node
	}

	nodes := cssxToNodes(value)

	if len(nodes) == 0 {
		return nil
	}

	return nodes[0]
}

func cssxNodesToAny(nodes []*html.Node) []any {
	out := make([]any, 0, len(nodes))

	for _, node := range nodes {
		if node != nil {
			out = append(out, node)
		}
	}

	return out
}

func cssxContainsNode(ancestor, node *html.Node) bool {
	if ancestor == nil || node == nil {
		return false
	}

	for cursor := node; cursor != nil; cursor = cursor.Parent {
		if cursor == ancestor {
			return true
		}
	}

	return false
}

func cssxDedupeNodes(nodes []*html.Node) []*html.Node {
	out := make([]*html.Node, 0, len(nodes))
	seen := make(map[*html.Node]struct{}, len(nodes))

	for _, node := range nodes {
		if node == nil {
			continue
		}

		if _, ok := seen[node]; ok {
			continue
		}

		seen[node] = struct{}{}
		out = append(out, node)
	}

	return out
}

func cssxTextOf(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case []any:
		var b strings.Builder

		for _, item := range v {
			b.WriteString(cssxTextOf(item))
		}

		return b.String()
	case *html.Node:
		return cssxTextContent(v)
	default:
		return fmt.Sprint(v)
	}
}

func cssxNormalizeSpace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func cssxTextContent(node *html.Node) string {
	if node == nil {
		return ""
	}

	var b strings.Builder
	var visit func(*html.Node)

	visit = func(current *html.Node) {
		if current.Type == html.TextNode {
			b.WriteString(current.Data)
		}

		for child := current.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}

	visit(node)

	return b.String()
}

func cssxOwnText(node *html.Node) string {
	if node == nil {
		return ""
	}

	var b strings.Builder

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode {
			b.WriteString(child.Data)
		}
	}

	return b.String()
}

func cssxParentElement(node *html.Node) *html.Node {
	if node == nil {
		return nil
	}

	parent := node.Parent

	if parent == nil || parent.Type != html.ElementNode {
		return nil
	}

	return parent
}

func cssxElementChildren(node *html.Node) []*html.Node {
	if node == nil {
		return nil
	}

	out := make([]*html.Node, 0)

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode {
			out = append(out, child)
		}
	}

	return out
}

func cssxNextElementSibling(node *html.Node) *html.Node {
	for sibling := node.NextSibling; sibling != nil; sibling = sibling.NextSibling {
		if sibling.Type == html.ElementNode {
			return sibling
		}
	}

	return nil
}

func cssxPrevElementSibling(node *html.Node) *html.Node {
	for sibling := node.PrevSibling; sibling != nil; sibling = sibling.PrevSibling {
		if sibling.Type == html.ElementNode {
			return sibling
		}
	}

	return nil
}

func cssxNodeAttr(node *html.Node, name string) (string, bool) {
	if node == nil || name == "" {
		return "", false
	}

	for _, attr := range node.Attr {
		if attr.Key == name {
			return attr.Val, true
		}
	}

	return "", false
}

func cssxNodeHasAttr(node *html.Node, name string) bool {
	_, ok := cssxNodeAttr(node, name)

	return ok
}

func cssxNodeProp(node *html.Node, name string) any {
	if node == nil {
		return nil
	}

	prop := strings.ToLower(strings.TrimSpace(name))

	switch prop {
	case "value":
		if value, ok := cssxNodeAttr(node, "value"); ok {
			return value
		}

		return nil
	case "checked", "selected", "disabled":
		return cssxNodeHasAttr(node, prop)
	default:
		if value, ok := cssxNodeAttr(node, prop); ok {
			return value
		}

		return nil
	}
}

func cssxInnerHTML(node *html.Node) any {
	if node == nil {
		return nil
	}

	var b bytes.Buffer

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if err := html.Render(&b, child); err != nil {
			return nil
		}
	}

	return b.String()
}

func cssxOuterHTML(node *html.Node) any {
	if node == nil {
		return nil
	}

	var b bytes.Buffer

	if err := html.Render(&b, node); err != nil {
		return nil
	}

	return b.String()
}

func cssxAsURL(value any, baseURL *url.URL) any {
	input := ""

	switch v := value.(type) {
	case nil:
		return nil
	case string:
		input = v
	default:
		input = fmt.Sprint(v)
	}

	input = strings.TrimSpace(input)

	if input == "" {
		return nil
	}

	resolved, err := cssxResolveURL(input, baseURL)

	if err != nil {
		return nil
	}

	return resolved.String()
}

func cssxParseURL(value any, baseURL *url.URL) any {
	input := ""

	switch v := value.(type) {
	case nil:
		return nil
	case string:
		input = v
	default:
		input = fmt.Sprint(v)
	}

	input = strings.TrimSpace(input)

	if input == "" {
		return nil
	}

	parsed, err := cssxResolveURL(input, baseURL)

	if err != nil {
		return nil
	}

	protocol := ""
	if parsed.Scheme != "" {
		protocol = parsed.Scheme + ":"
	}

	search := ""
	if parsed.RawQuery != "" {
		search = "?" + parsed.RawQuery
	}

	hash := ""
	if parsed.Fragment != "" {
		hash = "#" + parsed.Fragment
	}

	origin := ""
	if parsed.Scheme != "" && parsed.Host != "" {
		origin = parsed.Scheme + "://" + parsed.Host
	}

	password, _ := parsed.User.Password()

	return map[string]any{
		"href":     parsed.String(),
		"protocol": protocol,
		"username": parsed.User.Username(),
		"password": password,
		"host":     parsed.Host,
		"hostname": parsed.Hostname(),
		"port":     parsed.Port(),
		"pathname": parsed.EscapedPath(),
		"search":   search,
		"hash":     hash,
		"origin":   origin,
	}
}

func cssxResolveURL(raw string, baseURL *url.URL) (*url.URL, error) {
	parsed, err := url.Parse(raw)

	if err != nil {
		return nil, err
	}

	if baseURL == nil {
		if !parsed.IsAbs() {
			return nil, fmt.Errorf("relative URL without base")
		}

		return parsed, nil
	}

	return baseURL.ResolveReference(parsed), nil
}

func cssxToNumber(value any) any {
	str := strings.TrimSpace(cssxTextOf(value))

	if str == "" {
		return nil
	}

	normalized := cssxNonNumeric.ReplaceAllString(str, "")

	if normalized == "" {
		return nil
	}

	if strings.Contains(normalized, ".") {
		normalized = strings.ReplaceAll(normalized, ",", "")
	} else {
		commaCount := strings.Count(normalized, ",")

		switch {
		case commaCount > 1:
			normalized = strings.ReplaceAll(normalized, ",", "")
		case commaCount == 1:
			idx := strings.Index(normalized, ",")
			head := strings.TrimLeft(normalized[:idx], "+-")
			tail := normalized[idx+1:]

			if len(tail) == 3 && cssxIsDigits(head) && cssxIsDigits(tail) {
				normalized = strings.ReplaceAll(normalized, ",", "")
			} else {
				normalized = strings.Replace(normalized, ",", ".", 1)
			}
		}
	}

	out, err := strconv.ParseFloat(normalized, 64)

	if err != nil || math.IsNaN(out) || math.IsInf(out, 0) {
		return nil
	}

	return out
}

func cssxIsDigits(input string) bool {
	if input == "" {
		return false
	}

	for _, ch := range input {
		if ch < '0' || ch > '9' {
			return false
		}
	}

	return true
}

func cssxToDate(value any, args []any) any {
	str := strings.TrimSpace(cssxTextOf(value))

	if str == "" {
		return nil
	}

	layouts := []string{}

	if len(args) > 0 {
		if layout, ok := args[0].(string); ok && strings.TrimSpace(layout) != "" {
			layouts = append(layouts, layout)
		}
	}

	layouts = append(layouts,
		time.RFC3339Nano,
		time.RFC3339,
		time.DateTime,
		time.DateOnly,
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC850,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		"2006-01-02 15:04",
		"02 Jan 2006",
		"Jan 2, 2006",
	)

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, str)

		if err == nil {
			return parsed.UTC().Format(time.RFC3339Nano)
		}
	}

	return nil
}

func cssxAnyEqual(left, right any) bool {
	switch l := left.(type) {
	case nil:
		return right == nil
	case *html.Node:
		r, ok := right.(*html.Node)
		return ok && l == r
	case string:
		r, ok := right.(string)
		return ok && l == r
	case bool:
		r, ok := right.(bool)
		return ok && l == r
	case int:
		r, ok := right.(int)
		return ok && l == r
	case float64:
		r, ok := right.(float64)
		return ok && l == r
	default:
		return fmt.Sprint(left) == fmt.Sprint(right)
	}
}

func cssxArrayContains(items []any, target any) bool {
	for _, item := range items {
		if cssxAnyEqual(item, target) {
			return true
		}
	}

	return false
}

func cssxTruthy(value any) bool {
	switch v := value.(type) {
	case nil:
		return false
	case bool:
		return v
	case string:
		return v != ""
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0 && !math.IsNaN(v)
	default:
		return true
	}
}

func cssxSliceArray(items []any, start, end int) []any {
	length := len(items)

	if start < 0 {
		start += length
	}

	if end < 0 {
		end += length
	}

	if start < 0 {
		start = 0
	}

	if start > length {
		start = length
	}

	if end < 0 {
		end = 0
	}

	if end > length {
		end = length
	}

	if end < start {
		end = start
	}

	out := make([]any, end-start)
	copy(out, items[start:end])

	return out
}
