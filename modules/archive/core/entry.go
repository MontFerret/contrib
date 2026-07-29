package core

import (
	"io/fs"
	"time"
)

type (
	entryKind uint8

	// Entry is the public metadata returned by LIST.
	Entry struct {
		CompressedSize *int64  `json:"compressedSize"`
		ModTime        *string `json:"modTime"`
		LinkName       *string `json:"linkName"`
		Name           string  `json:"name"`
		Mode           string  `json:"mode"`
		Format         Format  `json:"format"`
		Size           int64   `json:"size"`
		fileMode       fs.FileMode
		IsDir          bool `json:"isDir"`
		IsRegular      bool `json:"isRegular"`
		IsSymlink      bool `json:"isSymlink"`
		kind           entryKind
	}

	// ExtractResult describes one eligible archive entry.
	ExtractResult struct {
		Path    *string `json:"path"`
		Reason  *string `json:"reason"`
		Name    string  `json:"name"`
		Size    int64   `json:"size"`
		IsDir   bool    `json:"isDir"`
		Skipped bool    `json:"skipped"`
	}
)

const (
	entryRegular entryKind = iota
	entryDirectory
	entrySymlink
	entryHardlink
	entrySpecial
)

func newEntry(
	name string,
	size int64,
	compressedSize *int64,
	mode fs.FileMode,
	modTime time.Time,
	linkName string,
	format Format,
	kind entryKind,
) Entry {
	var formattedTime *string

	if !modTime.IsZero() {
		value := modTime.UTC().Format(time.RFC3339)
		formattedTime = &value
	}

	var formattedLink *string
	if linkName != "" {
		formattedLink = &linkName
	}

	return Entry{
		Name:           name,
		Size:           size,
		CompressedSize: compressedSize,
		Mode:           formatMode(mode),
		ModTime:        formattedTime,
		IsDir:          kind == entryDirectory,
		IsRegular:      kind == entryRegular,
		IsSymlink:      kind == entrySymlink,
		LinkName:       formattedLink,
		Format:         format,
		kind:           kind,
		fileMode:       mode,
	}
}

func formatMode(mode fs.FileMode) string {
	const digits = "0000"

	value := mode.Perm()
	out := make([]byte, len(digits))

	for i := len(out) - 1; i >= 0; i-- {
		out[i] = byte('0' + value&7)
		value >>= 3
	}

	return string(out)
}
