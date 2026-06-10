package cssx

// Expression is a CSSX pseudo-function.
type Expression string

const (
	ExpressionFirst Expression = ":first"
	ExpressionLast  Expression = ":last"
	ExpressionNth   Expression = ":nth"

	ExpressionTake         Expression = ":take"
	ExpressionSkip         Expression = ":skip"
	ExpressionSlice        Expression = ":slice"
	ExpressionCompact      Expression = ":compact"
	ExpressionDistinct     Expression = ":distinct"
	ExpressionDedupeByAttr Expression = ":dedupeByAttr"
	ExpressionDedupeByText Expression = ":dedupeByText"

	ExpressionWithin   Expression = ":within"
	ExpressionHas      Expression = ":has"
	ExpressionMatches  Expression = ":matches"
	ExpressionNot      Expression = ":not"
	ExpressionWithAttr Expression = ":withAttr"
	ExpressionWithText Expression = ":withText"

	ExpressionParent   Expression = ":parent"
	ExpressionClosest  Expression = ":closest"
	ExpressionChildren Expression = ":children"
	ExpressionNext     Expression = ":next"
	ExpressionPrev     Expression = ":prev"
	ExpressionSiblings Expression = ":siblings"

	ExpressionText      Expression = ":text"
	ExpressionOwnText   Expression = ":ownText"
	ExpressionNormalize Expression = ":normalize"
	ExpressionTrim      Expression = ":trim"
	ExpressionAttr      Expression = ":attr"
	ExpressionProp      Expression = ":prop"
	ExpressionHTML      Expression = ":html"
	ExpressionOuterHTML Expression = ":outerHtml"
	ExpressionValue     Expression = ":value"
	ExpressionAbsURL    Expression = ":absUrl"
	ExpressionURL       Expression = ":url"
	ExpressionParseURL  Expression = ":parseUrl"
	ExpressionReplace   Expression = ":replace"
	ExpressionRegex     Expression = ":regex"
	ExpressionToNumber  Expression = ":toNumber"
	ExpressionToDate    Expression = ":toDate"

	ExpressionExists  Expression = ":exists"
	ExpressionEmpty   Expression = ":empty"
	ExpressionCount   Expression = ":count"
	ExpressionOne     Expression = ":one"
	ExpressionIndexOf Expression = ":indexOf"
	ExpressionLen     Expression = ":len"
	ExpressionJoin    Expression = ":join"
)
