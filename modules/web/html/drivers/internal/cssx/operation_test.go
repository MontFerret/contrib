package cssx

import "testing"

func TestOperationFamilies(t *testing.T) {
	t.Parallel()

	cases := map[Expression]OperationFamily{
		ExpressionText:     FamilyMap,
		ExpressionParent:   FamilyTraversal,
		ExpressionHas:      FamilyFilter,
		ExpressionCompact:  FamilySelection,
		ExpressionCount:    FamilyReducer,
		ExpressionFirst:    FamilyCardinality,
		ExpressionSiblings: FamilyTraversal,
		ExpressionDistinct: FamilySelection,
		ExpressionOne:      FamilyReducer,
	}

	for expression, expected := range cases {
		op, err := ResolveOperation(string(expression))
		if err != nil {
			t.Fatalf("resolve %s: %v", expression, err)
		}

		if op.Family != expected {
			t.Fatalf("%s: expected %s, got %s", expression, expected, op.Family)
		}
	}

	ops, err := CompileOps(`:text(p)`)
	if err != nil {
		t.Fatalf("compile mapped expression: %v", err)
	}

	if len(ops) != 2 || ops[1].Family != FamilyMap {
		t.Fatalf("expected compiled map family, got %#v", ops)
	}
}

func TestRemovedOperationsAreRejected(t *testing.T) {
	t.Parallel()

	for _, expression := range []string{`:texts(p)`, `:attrs("href", a)`, `:filter(a, a)`} {
		if _, err := CompileOps(expression); err == nil {
			t.Fatalf("expected %s to be rejected", expression)
		}
	}
}

func TestRelativeCriteriaValidation(t *testing.T) {
	t.Parallel()

	valid := []string{
		`.product >> :has(".price")`,
		`:matches(".active", li)`,
		`:closest(".card", .title)`,
		`.item >> :siblings()`,
	}
	for _, expression := range valid {
		if _, err := CompileOps(expression); err != nil {
			t.Fatalf("expected %s to compile: %v", expression, err)
		}
	}

	invalid := []string{
		`:has(.price, .product)`,
		`:matches(.active, li)`,
		`:closest(.card)`,
	}
	for _, expression := range invalid {
		if _, err := CompileOps(expression); err == nil {
			t.Fatalf("expected %s to be rejected", expression)
		}
	}
}
