package core

import (
	"fmt"
	"path"
	"strings"
)

type (
	plannedEntry struct {
		canonical   string
		destination string
		segments    []string
		entry       Entry
		index       int
		skipLink    bool
	}

	manifestBuilder struct {
		filter      *entryFilter
		seen        map[string]entryKind
		destination string
		links       string
		entries     []plannedEntry
		maxEntry    int64
	}
)

func newManifestBuilder(
	filter *entryFilter,
	destination string,
	links string,
	maxEntry int64,
) *manifestBuilder {
	return &manifestBuilder{
		filter:      filter,
		destination: destination,
		links:       links,
		maxEntry:    maxEntry,
		seen:        make(map[string]entryKind),
	}
}

func (b *manifestBuilder) add(index int, entry Entry) error {
	segments, canonical, err := validateEntryPath(entry.Name, entry.kind == entryDirectory)
	if err != nil {
		return err
	}

	if !b.filter.matches(canonical) {
		return nil
	}

	if _, exists := b.seen[canonical]; exists {
		return fmt.Errorf("archive contains duplicate extraction path %q", canonical)
	}

	b.seen[canonical] = entry.kind

	planned := plannedEntry{
		entry:       entry,
		segments:    segments,
		canonical:   canonical,
		destination: joinDestination(b.destination, segments),
		index:       index,
	}

	switch entry.kind {
	case entrySymlink, entryHardlink:
		if b.links == "error" {
			return fmt.Errorf("archive entry %q is a link and links are disabled", entry.Name)
		}

		planned.destination = ""
		planned.skipLink = true
	case entrySpecial:
		return fmt.Errorf("archive entry %q has an unsupported special file type", entry.Name)
	case entryRegular:
		if entry.Size > b.maxEntry {
			return &LimitError{Entry: entry.Name, Size: entry.Size, Limit: b.maxEntry}
		}
	}

	b.entries = append(b.entries, planned)

	return nil
}

func (b *manifestBuilder) build() ([]plannedEntry, error) {
	for _, planned := range b.entries {
		if planned.skipLink {
			continue
		}

		parent := path.Dir(planned.canonical)
		for parent != "." {
			if kind, exists := b.seen[parent]; exists && kind != entryDirectory {
				return nil, fmt.Errorf(
					"archive path %q requires non-directory entry %q as a parent",
					planned.canonical,
					parent,
				)
			}

			next := path.Dir(parent)

			if next == parent || strings.HasPrefix(parent, "../") {
				break
			}

			parent = next
		}
	}

	return append([]plannedEntry(nil), b.entries...), nil
}
