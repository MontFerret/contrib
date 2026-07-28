package core

import (
	"context"
	"errors"
	"io/fs"
	"testing"

	ferretfs "github.com/MontFerret/ferret/v2/pkg/fs"
)

func TestDestinationFSRemovesPartialFileAfterReadFailure(t *testing.T) {
	t.Parallel()

	filesystem, err := ferretfs.New(ferretfs.WithRoot(t.TempDir()))
	if err != nil {
		t.Fatalf("create filesystem: %v", err)
	}

	destination := newDestinationFS(filesystem, true, false)
	target := "dist/file.txt"
	planned := plannedEntry{
		entry: Entry{
			Name:     "file.txt",
			Size:     8,
			kind:     entryRegular,
			fileMode: 0o600,
		},
		canonical:   "file.txt",
		destination: target,
	}

	err = destination.writeFile(
		context.Background(),
		planned,
		&failingReader{data: []byte("partial")},
		64,
	)
	if !errors.Is(err, errSyntheticRead) {
		t.Fatalf("expected source read error, got %v", err)
	}
	if _, err := filesystem.Stat(target); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected partial file cleanup, got %v", err)
	}
}
