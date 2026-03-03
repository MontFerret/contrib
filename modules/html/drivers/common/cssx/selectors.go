package cssx

// Expression is a cssx "extended CSS" operation.
//
// In regular CSS, a selector string (e.g. ".section .item") can only select nodes.
// cssx extends CSS with Ferret-style pseudo functions (called "expressions") that
// can be nested and combined to post-process selection results.
//
// Expressions are written inside a cssx string using a leading ':' name and
// optional arguments:
//
//	:first(h1)                 // select nodes with CSS, then take first
//	:text(:first(h1))          // take first <h1>, then read its text
//	:count(.product)           // count matching nodes
//	:attr("href", :first(a))   // take first <a>, read href attribute
//
// Expressions are NOT CSS pseudo-classes like ":nth-child(2)".
// They are an extension layer evaluated by Ferret (e.g. compiled into a pipeline
// and executed in the browser via JS or in an in-memory DOM engine).
//
// Usage examples (Ferret-style):
//
//	LET nodes = el[~ css`:take(5)`]                    // -> list of nodes
//	LET title = el[~ css`:text(:first(h1))`]           // -> string
//	LET n = el[~ css`:count(.product)`]                // -> number
//	LET href = el[~ css`:attr("href", :first(a.cta))`] // -> string|null
//
// Note: expressions can return different kinds of values (node lists, a single node,
// scalars). Node-returning expressions are suitable for "query" operations; scalar-
// returning expressions require "eval" (or equivalent) depending on your runtime API.
type Expression string

var (
	// ExpressionFirst selects the first node from a node set, or the first node matching
	// a selector under the current root.
	//
	// Examples:
	//   :first(.item)          -> node|null
	//   :text(:first(h1))      -> string|null
	ExpressionFirst Expression = ":first"

	// ExpressionLast selects the last node from a node set (or last match for selector).
	//
	// Example:
	//   :last(.item)           -> node|null
	ExpressionLast Expression = ":last"

	// ExpressionNth selects the nth node from a node set (indexing convention is up to you:
	// 0-based or 1-based - document it and keep it consistent).
	//
	// Examples:
	//   :nth(2, .item)         -> node|null
	//   :text(:nth(0, h2))     -> string|null
	ExpressionNth Expression = ":nth"

	// ExpressionTake keeps only the first N nodes from a node set.
	//
	// Example:
	//   :take(5)               -> nodes (from the current node set)
	//   :take(5, .item)        -> nodes (first 5 matches of .item)
	ExpressionTake Expression = ":take"

	// ExpressionSkip skips the first N nodes from a node set.
	//
	// Example:
	//   :skip(10, .item)       -> nodes
	ExpressionSkip Expression = ":skip"

	// ExpressionSlice selects a subrange from a node set (offset, limit).
	//
	// Example:
	//   :slice(10, 5, .item)   -> nodes (items 10..14)
	ExpressionSlice Expression = ":slice"

	// ExpressionWithin evaluates a selector/expression within a scoped subtree.
	// Use this to avoid ambiguity when composing queries across multiple sections.
	//
	// Example:
	//   :within(.product, :text(:first(h2))) -> string|null
	ExpressionWithin Expression = ":within"

	// ExpressionParent returns the parent of the first node in the set.
	//
	// Example:
	//   :parent(:first(.item)) -> node|null
	ExpressionParent Expression = ":parent"

	// ExpressionClosest finds the closest ancestor (including self depending on your semantics)
	// matching a selector, starting from the first node in the set.
	//
	// Example:
	//   :closest(.card, :first(.title)) -> node|null
	ExpressionClosest Expression = ":closest"

	// ExpressionChildren returns children of the first node (optionally filtered by selector).
	//
	// Example:
	//   :children(li, :first(ul.menu)) -> nodes
	ExpressionChildren Expression = ":children"

	// ExpressionNext returns the next sibling of the first node (optionally filtered by selector).
	//
	// Example:
	//   :next(.value, :first(.label)) -> node|null
	ExpressionNext Expression = ":next"

	// ExpressionPrev returns the previous sibling of the first node (optionally filtered by selector).
	//
	// Example:
	//   :prev(.label, :first(.value)) -> node|null
	ExpressionPrev Expression = ":prev"

	// ExpressionExists returns true if the selector/expression yields at least one node.
	//
	// Example:
	//   :exists(form#captcha)  -> bool
	ExpressionExists Expression = ":exists"

	// ExpressionEmpty returns true if the selector/expression yields zero nodes.
	//
	// Example:
	//   :empty(.results)       -> bool
	ExpressionEmpty Expression = ":empty"

	// ExpressionHas returns true if the first node in the set has a descendant matching selector.
	//
	// Example:
	//   :has(.price, :first(.product)) -> bool
	ExpressionHas Expression = ":has"

	// ExpressionMatches returns true if the first node in the set matches the selector.
	//
	// Example:
	//   :matches(.active, :first(li)) -> bool
	ExpressionMatches Expression = ":matches"

	// ExpressionCount returns the number of nodes produced by selector/expression.
	//
	// Examples:
	//   :count(.item)          -> number
	//   :count(:take(5, .item))-> number (5)
	ExpressionCount Expression = ":count"

	// ExpressionIndexOf returns index of a target node within a node set (optional / advanced).
	//
	// Example:
	//   :indexOf(.item, :first(.selected)) -> number
	ExpressionIndexOf Expression = ":indexOf"

	// ExpressionLen returns length of a list or string (optional convenience alias).
	//
	// Examples:
	//   :len(:texts(.tag))     -> number
	//   :len(:text(h1))        -> number
	ExpressionLen Expression = ":len"

	// ExpressionText extracts textContent from the first node of selector/expression.
	//
	// Example:
	//   :text(:first(h1))      -> string|null
	ExpressionText Expression = ":text"

	// ExpressionTexts extracts textContent from all nodes of selector/expression.
	//
	// Example:
	//   :texts(.tag)           -> []string
	ExpressionTexts Expression = ":texts"

	// ExpressionOwnText extracts only direct text nodes (no descendant text).
	//
	// Example:
	//   :ownText(.price)       -> string|null
	ExpressionOwnText Expression = ":ownText"

	// ExpressionNormalize normalizes whitespace (trim + collapse internal whitespace).
	//
	// Example:
	//   :normalize(:text(h1))  -> string
	ExpressionNormalize Expression = ":normalize"

	// ExpressionTrim trims leading/trailing whitespace.
	//
	// Example:
	//   :trim(:text(h1))       -> string
	ExpressionTrim Expression = ":trim"

	// ExpressionJoin joins a list of strings using a separator.
	//
	// Example:
	//   :join(", ", :texts(.tag)) -> string
	ExpressionJoin Expression = ":join"

	// ExpressionAttr reads an attribute from the first node.
	//
	// Example:
	//   :attr("href", :first(a)) -> string|null
	ExpressionAttr Expression = ":attr"

	// ExpressionAttrs reads an attribute from all nodes.
	//
	// Example:
	//   :attrs("href", a)      -> []string
	ExpressionAttrs Expression = ":attrs"

	// ExpressionProp reads a DOM property from the first node (value, checked, etc.).
	//
	// Example:
	//   :prop("value", :first(input[name="q"])) -> any
	ExpressionProp Expression = ":prop"

	// ExpressionHTML returns innerHTML of the first node.
	//
	// Example:
	//   :html(:first(.content)) -> string|null
	ExpressionHTML Expression = ":html"

	// ExpressionOuterHTML returns outerHTML of the first node.
	//
	// Example:
	//   :outerHtml(:first(.content)) -> string|null
	ExpressionOuterHTML Expression = ":outerHtml"

	// ExpressionValue is a convenience for form-field value (often maps to property "value").
	//
	// Example:
	//   :value(:first(input[name="q"])) -> string|null
	ExpressionValue Expression = ":value"

	// ExpressionAbsURL resolves a relative URL (usually from href/src) into an absolute URL,
	// using document.baseURI / page URL (execution-time detail).
	//
	// Example:
	//   :absUrl(:attr("href", :first(a))) -> string|null
	ExpressionAbsURL Expression = ":absUrl"

	// ExpressionURL reads a URL from a named attribute and resolves it to absolute.
	//
	// Example:
	//   :url("href", :first(a)) -> string|null
	ExpressionURL Expression = ":url"

	// ExpressionParseURL parses a URL string into a structured object (optional).
	//
	// Example:
	//   :parseUrl(:url("href", :first(a))) -> object
	ExpressionParseURL Expression = ":parseUrl"

	// ExpressionFilter filters a node set (advanced; exact predicate language is up to your runtime).
	//
	// Example (one possible meaning):
	//   :filter(:has(.price), .product) -> nodes
	ExpressionFilter Expression = ":filter"

	// ExpressionWithAttr keeps nodes that have a given attribute.
	//
	// Example:
	//   :withAttr("href", a)   -> nodes
	ExpressionWithAttr Expression = ":withAttr"

	// ExpressionWithText keeps nodes whose text matches a substring/regex (advanced).
	//
	// Example:
	//   :withText("Sale", .product) -> nodes
	ExpressionWithText Expression = ":withText"

	// ExpressionDedupeByAttr deduplicates a node set by attribute value.
	//
	// Example:
	//   :dedupeByAttr("href", a) -> nodes
	ExpressionDedupeByAttr Expression = ":dedupeByAttr"

	// ExpressionDedupeByText deduplicates a node set by normalized text.
	//
	// Example:
	//   :dedupeByText(.tag)    -> nodes
	ExpressionDedupeByText Expression = ":dedupeByText"

	// ExpressionReplace replaces text using a pattern (regex or substring; define at runtime).
	//
	// Example:
	//   :replace("\\s+", " ", :text(h1)) -> string
	ExpressionReplace Expression = ":replace"

	// ExpressionRegex extracts a regex match/group from a string.
	//
	// Example:
	//   :regex("(\\d+(?:\\.\\d+)?)", 1, :text(.price)) -> string|null
	ExpressionRegex Expression = ":regex"

	// ExpressionToNumber converts a string to number (after cleanup).
	//
	// Example:
	//   :toNumber(:regex("(\\d+)", 1, :text(.price))) -> number|null
	ExpressionToNumber Expression = ":toNumber"

	// ExpressionToDate converts a string to date/time using a layout/hint (optional).
	//
	// Example:
	//   :toDate("2006-01-02", :text(time)) -> date|null
	ExpressionToDate Expression = ":toDate"
)

var selectorLookup = map[string]Expression{
	string(ExpressionFirst):        ExpressionFirst,
	string(ExpressionLast):         ExpressionLast,
	string(ExpressionNth):          ExpressionNth,
	string(ExpressionTake):         ExpressionTake,
	string(ExpressionSkip):         ExpressionSkip,
	string(ExpressionSlice):        ExpressionSlice,
	string(ExpressionWithin):       ExpressionWithin,
	string(ExpressionParent):       ExpressionParent,
	string(ExpressionClosest):      ExpressionClosest,
	string(ExpressionChildren):     ExpressionChildren,
	string(ExpressionNext):         ExpressionNext,
	string(ExpressionPrev):         ExpressionPrev,
	string(ExpressionExists):       ExpressionExists,
	string(ExpressionEmpty):        ExpressionEmpty,
	string(ExpressionHas):          ExpressionHas,
	string(ExpressionMatches):      ExpressionMatches,
	string(ExpressionCount):        ExpressionCount,
	string(ExpressionIndexOf):      ExpressionIndexOf,
	string(ExpressionLen):          ExpressionLen,
	string(ExpressionText):         ExpressionText,
	string(ExpressionTexts):        ExpressionTexts,
	string(ExpressionOwnText):      ExpressionOwnText,
	string(ExpressionNormalize):    ExpressionNormalize,
	string(ExpressionTrim):         ExpressionTrim,
	string(ExpressionJoin):         ExpressionJoin,
	string(ExpressionAttr):         ExpressionAttr,
	string(ExpressionAttrs):        ExpressionAttrs,
	string(ExpressionProp):         ExpressionProp,
	string(ExpressionHTML):         ExpressionHTML,
	string(ExpressionOuterHTML):    ExpressionOuterHTML,
	string(ExpressionValue):        ExpressionValue,
	string(ExpressionAbsURL):       ExpressionAbsURL,
	string(ExpressionURL):          ExpressionURL,
	string(ExpressionParseURL):     ExpressionParseURL,
	string(ExpressionFilter):       ExpressionFilter,
	string(ExpressionWithAttr):     ExpressionWithAttr,
	string(ExpressionWithText):     ExpressionWithText,
	string(ExpressionDedupeByAttr): ExpressionDedupeByAttr,
	string(ExpressionDedupeByText): ExpressionDedupeByText,
	string(ExpressionReplace):      ExpressionReplace,
	string(ExpressionRegex):        ExpressionRegex,
	string(ExpressionToNumber):     ExpressionToNumber,
	string(ExpressionToDate):       ExpressionToDate,
}
