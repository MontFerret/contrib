package memory

import (
	"bytes"
	"fmt"
	"math"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"

	"github.com/MontFerret/contrib/modules/web/html/drivers/internal/cssx"
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

	op, err := cssx.ResolveOperation(string(name))
	if err != nil {
		return []any{}
	}

	switch op.Family {
	case cssx.FamilyCardinality:
		return cssxApplyCardinality(name, args, input)
	case cssx.FamilySelection:
		return cssxApplySelection(name, args, input)
	case cssx.FamilyTraversal:
		return cssxApplyTraversal(name, args, input)
	case cssx.FamilyFilter:
		return cssxApplyFilter(name, args, input)
	case cssx.FamilyMap:
		return cssxApplyMap(name, args, input, baseURL)
	case cssx.FamilyReducer:
		return cssxApplyReducer(name, args, values, input)
	default:
		return []any{}
	}
}

func cssxApplyCardinality(name cssx.Expression, args []any, input any) any {
	items := cssxToArray(input)

	switch name {
	case cssx.ExpressionFirst:
		if len(items) > 0 {
			return items[0]
		}
	case cssx.ExpressionLast:
		if len(items) > 0 {
			return items[len(items)-1]
		}
	case cssx.ExpressionNth:
		idx, ok := cssxToInt(cssxArgNumber(args, 0))
		if ok && idx >= 0 && idx < len(items) {
			return items[idx]
		}
	}

	return nil
}

func cssxApplySelection(name cssx.Expression, args []any, input any) []any {
	items := cssxToArray(input)

	switch name {
	case cssx.ExpressionTake:
		n, ok := cssxToInt(cssxArgNumber(args, 0))
		if !ok || n <= 0 {
			return []any{}
		}
		if n > len(items) {
			n = len(items)
		}
		return append([]any(nil), items[:n]...)
	case cssx.ExpressionSkip:
		n, ok := cssxToInt(cssxArgNumber(args, 0))
		if !ok || n <= 0 {
			return append([]any(nil), items...)
		}
		if n >= len(items) {
			return []any{}
		}
		return append([]any(nil), items[n:]...)
	case cssx.ExpressionSlice:
		start, startOK := cssxToInt(cssxArgNumber(args, 0))
		count, countOK := cssxToInt(cssxArgNumber(args, 1))
		if !startOK || !countOK || count <= 0 {
			return []any{}
		}
		return cssxSliceArray(items, start, start+count)
	case cssx.ExpressionCompact:
		out := make([]any, 0, len(items))
		for _, item := range items {
			if item != nil {
				out = append(out, item)
			}
		}
		return out
	case cssx.ExpressionDistinct:
		out := make([]any, 0, len(items))
		for _, item := range items {
			if !cssxArrayContainsDistinct(out, item) {
				out = append(out, item)
			}
		}
		return out
	case cssx.ExpressionDedupeByAttr:
		seen := make(map[string]struct{}, len(items))
		out := make([]any, 0, len(items))
		for _, node := range cssxToNodes(items) {
			value, ok := cssxNodeAttr(node, cssxArgString(args, 0))
			key := "<none>"
			if ok {
				key = value
			}
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, node)
		}
		return out
	case cssx.ExpressionDedupeByText:
		seen := make(map[string]struct{}, len(items))
		out := make([]any, 0, len(items))
		for _, node := range cssxToNodes(items) {
			key := cssxNormalizeSpace(cssxTextContent(node))
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, node)
		}
		return out
	default:
		return []any{}
	}
}

func cssxApplyTraversal(name cssx.Expression, args []any, input any) []any {
	criterion := cssxArgString(args, 0)
	out := make([]any, 0)

	for _, node := range cssxToNodes(input) {
		switch name {
		case cssx.ExpressionParent:
			cssxAppendMatchingNode(&out, cssxParentElement(node), criterion)
		case cssx.ExpressionClosest:
			for current := node; current != nil; current = cssxParentElement(current) {
				if cssxNodeMatches(current, criterion) {
					out = append(out, current)
					break
				}
			}
		case cssx.ExpressionChildren:
			for _, child := range cssxElementChildren(node) {
				cssxAppendMatchingNode(&out, child, criterion)
			}
		case cssx.ExpressionNext:
			cssxAppendMatchingNode(&out, cssxNextElementSibling(node), criterion)
		case cssx.ExpressionPrev:
			cssxAppendMatchingNode(&out, cssxPrevElementSibling(node), criterion)
		case cssx.ExpressionSiblings:
			for _, sibling := range cssxElementSiblings(node) {
				cssxAppendMatchingNode(&out, sibling, criterion)
			}
		}
	}

	return out
}

func cssxApplyFilter(name cssx.Expression, args []any, input any) []any {
	criterion := cssxArgString(args, 0)
	out := make([]any, 0)

	if !cssxValidSelector(criterion) {
		return out
	}

	for _, node := range cssxToNodes(input) {
		keep := false

		switch name {
		case cssx.ExpressionWithin:
			keep = cssxNodeWithin(node, criterion)
		case cssx.ExpressionHas:
			keep = cssxNodeHas(node, criterion)
		case cssx.ExpressionMatches:
			keep = cssxNodeMatches(node, criterion)
		case cssx.ExpressionNot:
			keep = !cssxNodeMatches(node, criterion)
		case cssx.ExpressionWithAttr:
			keep = cssxNodeHasAttr(node, criterion)
		case cssx.ExpressionWithText:
			keep = strings.Contains(cssxTextContent(node), criterion)
		}

		if keep {
			out = append(out, node)
		}
	}

	return out
}

func cssxApplyMap(name cssx.Expression, args []any, input any, baseURL *url.URL) []any {
	items := cssxToArray(input)
	out := make([]any, 0, len(items))

	for _, item := range items {
		if item == nil {
			out = append(out, nil)
			continue
		}

		out = append(out, cssxMapItem(name, args, item, baseURL))
	}

	return out
}

func cssxMapItem(name cssx.Expression, args []any, input any, baseURL *url.URL) any {
	node, _ := input.(*html.Node)

	switch name {
	case cssx.ExpressionText:
		if node != nil {
			return cssxTextContent(node)
		}
	case cssx.ExpressionOwnText:
		if node != nil {
			return cssxOwnText(node)
		}
	case cssx.ExpressionNormalize:
		return cssxNormalizeSpace(cssxTextOf(input))
	case cssx.ExpressionTrim:
		return strings.TrimSpace(cssxTextOf(input))
	case cssx.ExpressionAttr:
		if value, ok := cssxNodeAttr(node, cssxArgString(args, 0)); ok {
			return value
		}
	case cssx.ExpressionProp:
		return cssxNodeProp(node, cssxArgString(args, 0))
	case cssx.ExpressionHTML:
		return cssxInnerHTML(node)
	case cssx.ExpressionOuterHTML:
		return cssxOuterHTML(node)
	case cssx.ExpressionValue:
		return cssxNodeProp(node, "value")
	case cssx.ExpressionAbsURL:
		return cssxAsURL(input, baseURL)
	case cssx.ExpressionURL:
		if value, ok := cssxNodeAttr(node, cssxArgString(args, 0)); ok {
			return cssxAsURL(value, baseURL)
		}
	case cssx.ExpressionParseURL:
		return cssxParseURL(input, baseURL)
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
		rx, err := regexp.Compile(pattern)
		if err != nil {
			return nil
		}
		matches := rx.FindStringSubmatch(cssxTextOf(input))
		if len(matches) > 0 && group >= 0 && group < len(matches) {
			return matches[group]
		}
	case cssx.ExpressionToNumber:
		return cssxToNumber(input)
	case cssx.ExpressionToDate:
		return cssxToDate(input, args)
	}

	return nil
}

func cssxApplyReducer(name cssx.Expression, args []any, values []any, input any) any {
	items := cssxToArray(input)

	switch name {
	case cssx.ExpressionExists:
		return len(items) > 0
	case cssx.ExpressionEmpty:
		return len(items) == 0
	case cssx.ExpressionCount:
		return len(items)
	case cssx.ExpressionOne:
		return len(items) == 1
	case cssx.ExpressionIndexOf:
		if len(values) < 2 {
			return -1
		}
		list := cssxToArray(values[0])
		target := cssxToArray(values[1])
		if len(target) == 0 {
			return -1
		}
		for idx, candidate := range list {
			if cssxAnyEqual(candidate, target[0]) {
				return idx
			}
		}
		return -1
	case cssx.ExpressionLen:
		if value, ok := input.(string); ok {
			return len(value)
		}
		return len(items)
	case cssx.ExpressionJoin:
		values := make([]string, 0, len(items))
		for _, item := range items {
			values = append(values, cssxTextOf(item))
		}
		return strings.Join(values, cssxArgString(args, 0))
	default:
		return nil
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
		out := make([]any, len(v))
		copy(out, v)
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

func cssxAppendMatchingNode(out *[]any, node *html.Node, criterion string) {
	if node == nil {
		return
	}

	if criterion != "" && !cssxNodeMatches(node, criterion) {
		return
	}

	*out = append(*out, node)
}

func cssxNodeMatches(node *html.Node, selector string) bool {
	if node == nil || selector == "" {
		return false
	}

	return goquery.NewDocumentFromNode(node).Selection.Is(selector)
}

func cssxValidSelector(selector string) bool {
	if selector == "" {
		return false
	}

	_, err := cascadia.Compile(selector)

	return err == nil
}

func cssxNodeHas(node *html.Node, selector string) bool {
	if node == nil || selector == "" {
		return false
	}

	return goquery.NewDocumentFromNode(node).Selection.Find(selector).Length() > 0
}

func cssxNodeWithin(node *html.Node, selector string) bool {
	for parent := cssxParentElement(node); parent != nil; parent = cssxParentElement(parent) {
		if cssxNodeMatches(parent, selector) {
			return true
		}
	}

	return false
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

func cssxElementSiblings(node *html.Node) []*html.Node {
	if node == nil || node.Parent == nil {
		return nil
	}

	out := make([]*html.Node, 0)

	for sibling := node.Parent.FirstChild; sibling != nil; sibling = sibling.NextSibling {
		if sibling != node && sibling.Type == html.ElementNode {
			out = append(out, sibling)
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

func cssxArrayContainsDistinct(items []any, target any) bool {
	for _, item := range items {
		if cssxDistinctEqual(item, target) {
			return true
		}
	}

	return false
}

func cssxDistinctEqual(left, right any) bool {
	switch left.(type) {
	case nil, string, bool, int, int64, float32, float64:
		return cssxAnyEqual(left, right)
	case *html.Node:
		return cssxAnyEqual(left, right)
	}

	leftValue := reflect.ValueOf(left)
	rightValue := reflect.ValueOf(right)

	return leftValue.IsValid() &&
		rightValue.IsValid() &&
		leftValue.Type() == rightValue.Type() &&
		leftValue == rightValue
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
