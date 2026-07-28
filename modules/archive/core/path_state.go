package core

import (
	"path"
)

func sandboxPrefixes(name string) []string {
	cleaned := path.Clean(name)
	if cleaned == "." || cleaned == "/" {
		return []string{cleaned}
	}

	var reverse []string
	for current := cleaned; current != "." && current != "/" && current != ""; current = path.Dir(current) {
		reverse = append(reverse, current)
		next := path.Dir(current)

		if next == current {
			break
		}
	}

	prefixes := make([]string, len(reverse))

	for i := range reverse {
		prefixes[len(reverse)-1-i] = reverse[i]
	}

	return prefixes
}
