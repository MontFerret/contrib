package core

import (
	"path"
	"strings"
	"unicode/utf8"
)

func validateEntryPath(name string, isDir bool) ([]string, string, error) {
	if name == "" {
		return nil, "", &EntryPathError{Name: name, Reason: "path is empty"}
	}

	if !utf8.ValidString(name) {
		return nil, "", &EntryPathError{Name: name, Reason: "path is not valid UTF-8"}
	}

	if strings.ContainsRune(name, '\x00') {
		return nil, "", &EntryPathError{Name: name, Reason: "path contains a NUL byte"}
	}

	if strings.ContainsRune(name, '\\') {
		return nil, "", &EntryPathError{Name: name, Reason: "backslashes are not allowed"}
	}

	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, "//") {
		return nil, "", &EntryPathError{Name: name, Reason: "absolute and UNC paths are not allowed"}
	}

	if len(name) >= 2 && isASCIIAlpha(name[0]) && name[1] == ':' {
		return nil, "", &EntryPathError{Name: name, Reason: "drive-qualified paths are not allowed"}
	}

	candidate := name
	if strings.HasSuffix(candidate, "/") {
		if !isDir {
			return nil, "", &EntryPathError{Name: name, Reason: "only directory entries may end with a slash"}
		}

		candidate = strings.TrimSuffix(candidate, "/")
	}

	if candidate == "" || candidate == "." {
		return nil, "", &EntryPathError{Name: name, Reason: "path does not identify an entry"}
	}

	segments := strings.Split(candidate, "/")

	for _, segment := range segments {
		switch segment {
		case "":
			return nil, "", &EntryPathError{Name: name, Reason: "empty path segments are not allowed"}
		case ".":
			return nil, "", &EntryPathError{Name: name, Reason: "dot path segments are not allowed"}
		case "..":
			return nil, "", &EntryPathError{Name: name, Reason: "parent path segments are not allowed"}
		}
	}

	canonical := strings.Join(segments, "/")
	if cleaned := path.Clean(canonical); cleaned != canonical || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return nil, "", &EntryPathError{Name: name, Reason: "path normalization is unsafe"}
	}

	return segments, canonical, nil
}

func isASCIIAlpha(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z'
}

func joinDestination(destination string, segments []string) string {
	parts := make([]string, 0, len(segments)+1)
	parts = append(parts, destination)
	parts = append(parts, segments...)

	return path.Join(parts...)
}
