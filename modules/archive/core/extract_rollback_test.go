package core

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ferretfs "github.com/MontFerret/ferret/v2/pkg/fs"
)

func TestExtractRollsBackEarlierOutputsAfterLateFailure(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "dist"), 0o755); err != nil {
		t.Fatalf("create existing destination: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "dist", "keep.txt"), []byte("keep"), 0o600); err != nil {
		t.Fatalf("write existing destination file: %v", err)
	}

	filesystem, err := ferretfs.New(ferretfs.WithRoot(root))
	if err != nil {
		t.Fatalf("create filesystem: %v", err)
	}

	preflight := tarBytes(t, []tarTestEntry{
		{name: "nested/first.txt", body: "first"},
		{name: "second.txt", body: "second"},
	})
	execution := tarBytes(t, []tarTestEntry{
		{name: "nested/first.txt", body: "first"},
		{name: "second.txt", body: "changed"},
	})
	changing := newChangingFS(filesystem, preflight, execution)
	ctx := ferretfs.WithFileSystem(context.Background(), changing)

	_, err = Extract(
		ctx,
		"archive.tar",
		"dist",
		ExtractConfig{
			Config:     DefaultConfig(),
			Format:     FormatTAR,
			CreateDirs: true,
			Links:      "skip",
		},
	)

	if err == nil || !strings.Contains(err.Error(), "source changed") {
		t.Fatalf("expected source-change failure, got %v", err)
	}

	if _, statErr := filesystem.Stat("dist/nested"); !errors.Is(statErr, fs.ErrNotExist) {
		t.Fatalf("expected created directory rollback, got %v", statErr)
	}

	content, readErr := filesystem.ReadFile("dist/keep.txt")

	if readErr != nil {
		t.Fatalf("read preserved destination file: %v", readErr)
	}

	if string(content) != "keep" {
		t.Fatalf("expected existing destination file to remain unchanged, got %q", content)
	}
}

type tarTestEntry struct {
	name string
	body string
}

func tarBytes(t *testing.T, entries []tarTestEntry) []byte {
	t.Helper()

	var buffer bytes.Buffer
	writer := tar.NewWriter(&buffer)

	for _, entry := range entries {
		header := &tar.Header{
			Name: entry.name,
			Mode: 0o600,
			Size: int64(len(entry.body)),
		}

		if err := writer.WriteHeader(header); err != nil {
			t.Fatalf("write TAR header: %v", err)
		}

		if _, err := writer.Write([]byte(entry.body)); err != nil {
			t.Fatalf("write TAR content: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close TAR fixture: %v", err)
	}

	return buffer.Bytes()
}
