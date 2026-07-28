package core

import "testing"

func TestEntryFilterUsesPathMatchAndExcludeWins(t *testing.T) {
	t.Parallel()

	filter, err := newEntryFilter(
		[]string{"docs/*"},
		[]string{"docs/*.map"},
	)
	if err != nil {
		t.Fatalf("create filter: %v", err)
	}

	if !filter.matches("docs/readme.txt") {
		t.Fatal("expected included file")
	}
	if filter.matches("docs/app.map") {
		t.Fatal("expected exclude to win")
	}
	if filter.matches("docs/nested/readme.txt") {
		t.Fatal("expected star not to cross a slash")
	}
}

func TestEntryFilterRejectsInvalidPattern(t *testing.T) {
	t.Parallel()

	if _, err := newEntryFilter([]string{"[bad"}, nil); err == nil {
		t.Fatal("expected invalid pattern error")
	}
}
