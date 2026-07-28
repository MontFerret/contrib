package core

import (
	"strings"
	"testing"
)

func TestValidateEntryPathRejectsUnsafeNames(t *testing.T) {
	t.Parallel()

	tests := []string{
		"",
		".",
		"../evil.txt",
		"../../evil.txt",
		"/absolute.txt",
		`C:\evil.txt`,
		"C:/evil.txt",
		`\\server\share\evil.txt`,
		"//server/share/evil.txt",
		"dir/../../evil.txt",
		"dir//evil.txt",
		"dir/./evil.txt",
		"dir\\..\\evil.txt",
		"safe-prefix/../evil.txt",
		"nul\x00name",
	}

	for _, name := range tests {
		name := name
		t.Run(strings.ReplaceAll(name, "/", "_"), func(t *testing.T) {
			t.Parallel()

			if _, _, err := validateEntryPath(name, false); err == nil {
				t.Fatalf("expected %q to be rejected", name)
			}
		})
	}
}

func TestValidateEntryPathCanonicalizesDirectory(t *testing.T) {
	t.Parallel()

	segments, canonical, err := validateEntryPath("docs/assets/", true)
	if err != nil {
		t.Fatalf("validate safe directory: %v", err)
	}
	if canonical != "docs/assets" || len(segments) != 2 {
		t.Fatalf("unexpected canonical path %q and segments %#v", canonical, segments)
	}
}
