package cssx

import (
	"errors"
	"fmt"
	"strings"
)

// OperationFamily defines how a CSSX pseudo-function treats its input selection.
type OperationFamily string

const (
	FamilyMap         OperationFamily = "map"
	FamilyTraversal   OperationFamily = "traversal"
	FamilyFilter      OperationFamily = "filter"
	FamilySelection   OperationFamily = "selection"
	FamilyReducer     OperationFamily = "reducer"
	FamilyCardinality OperationFamily = "cardinality"
)

// Operation describes the selection behavior of a CSSX pseudo-function.
type Operation struct {
	Expression Expression
	Family     OperationFamily
}

var operationLookup = map[string]Operation{
	string(ExpressionFirst): {ExpressionFirst, FamilyCardinality},
	string(ExpressionLast):  {ExpressionLast, FamilyCardinality},
	string(ExpressionNth):   {ExpressionNth, FamilyCardinality},

	string(ExpressionTake):         {ExpressionTake, FamilySelection},
	string(ExpressionSkip):         {ExpressionSkip, FamilySelection},
	string(ExpressionSlice):        {ExpressionSlice, FamilySelection},
	string(ExpressionCompact):      {ExpressionCompact, FamilySelection},
	string(ExpressionDistinct):     {ExpressionDistinct, FamilySelection},
	string(ExpressionDedupeByAttr): {ExpressionDedupeByAttr, FamilySelection},
	string(ExpressionDedupeByText): {ExpressionDedupeByText, FamilySelection},

	string(ExpressionWithin):   {ExpressionWithin, FamilyFilter},
	string(ExpressionHas):      {ExpressionHas, FamilyFilter},
	string(ExpressionMatches):  {ExpressionMatches, FamilyFilter},
	string(ExpressionNot):      {ExpressionNot, FamilyFilter},
	string(ExpressionWithAttr): {ExpressionWithAttr, FamilyFilter},
	string(ExpressionWithText): {ExpressionWithText, FamilyFilter},

	string(ExpressionParent):   {ExpressionParent, FamilyTraversal},
	string(ExpressionClosest):  {ExpressionClosest, FamilyTraversal},
	string(ExpressionChildren): {ExpressionChildren, FamilyTraversal},
	string(ExpressionNext):     {ExpressionNext, FamilyTraversal},
	string(ExpressionPrev):     {ExpressionPrev, FamilyTraversal},
	string(ExpressionSiblings): {ExpressionSiblings, FamilyTraversal},

	string(ExpressionText):      {ExpressionText, FamilyMap},
	string(ExpressionOwnText):   {ExpressionOwnText, FamilyMap},
	string(ExpressionNormalize): {ExpressionNormalize, FamilyMap},
	string(ExpressionTrim):      {ExpressionTrim, FamilyMap},
	string(ExpressionAttr):      {ExpressionAttr, FamilyMap},
	string(ExpressionProp):      {ExpressionProp, FamilyMap},
	string(ExpressionHTML):      {ExpressionHTML, FamilyMap},
	string(ExpressionOuterHTML): {ExpressionOuterHTML, FamilyMap},
	string(ExpressionValue):     {ExpressionValue, FamilyMap},
	string(ExpressionAbsURL):    {ExpressionAbsURL, FamilyMap},
	string(ExpressionURL):       {ExpressionURL, FamilyMap},
	string(ExpressionParseURL):  {ExpressionParseURL, FamilyMap},
	string(ExpressionReplace):   {ExpressionReplace, FamilyMap},
	string(ExpressionRegex):     {ExpressionRegex, FamilyMap},
	string(ExpressionToNumber):  {ExpressionToNumber, FamilyMap},
	string(ExpressionToDate):    {ExpressionToDate, FamilyMap},

	string(ExpressionExists):  {ExpressionExists, FamilyReducer},
	string(ExpressionEmpty):   {ExpressionEmpty, FamilyReducer},
	string(ExpressionCount):   {ExpressionCount, FamilyReducer},
	string(ExpressionOne):     {ExpressionOne, FamilyReducer},
	string(ExpressionIndexOf): {ExpressionIndexOf, FamilyReducer},
	string(ExpressionLen):     {ExpressionLen, FamilyReducer},
	string(ExpressionJoin):    {ExpressionJoin, FamilyReducer},
}

func ResolveOperation(selector string) (Operation, error) {
	value := strings.TrimSpace(selector)
	if value == "" {
		return Operation{}, errors.New("selector is empty")
	}

	if !strings.HasPrefix(value, ":") {
		value = ":" + value
	}

	resolved, ok := operationLookup[value]
	if !ok {
		return Operation{}, fmt.Errorf("unknown selector %q", value)
	}

	return resolved, nil
}

func ResolveSelector(selector string) (Expression, error) {
	op, err := ResolveOperation(selector)
	if err != nil {
		return "", err
	}

	return op.Expression, nil
}
