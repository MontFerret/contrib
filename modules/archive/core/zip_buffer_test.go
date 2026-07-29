package core

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"testing"

	ferretfs "github.com/MontFerret/ferret/v2/pkg/fs"
)

func TestListBuffersSequentialZIPWithinLimit(t *testing.T) {
	t.Parallel()

	data := zipBytesForTest(t, "file.txt", "content")
	filesystem := &sequentialFS{data: data}
	ctx := ferretfs.WithFileSystem(context.Background(), filesystem)

	entries, err := List(ctx, "archive.zip", DefaultListOptions(Config{
		MaxEntrySize:     1024,
		MaxZIPBufferSize: int64(len(data)),
	}))
	if err != nil {
		t.Fatalf("list sequential ZIP: %v", err)
	}

	if len(entries) != 1 || entries[0].Name != "file.txt" {
		t.Fatalf("unexpected entries %#v", entries)
	}

	if filesystem.last == nil || !filesystem.last.closed {
		t.Fatal("expected buffered source to close")
	}
}

func TestListRejectsSequentialZIPOverBufferLimit(t *testing.T) {
	t.Parallel()

	data := zipBytesForTest(t, "file.txt", "content")
	filesystem := &sequentialFS{data: data}
	ctx := ferretfs.WithFileSystem(context.Background(), filesystem)

	_, err := List(ctx, "archive.zip", DefaultListOptions(Config{
		MaxEntrySize:     1024,
		MaxZIPBufferSize: int64(len(data) - 1),
	}))

	var limitErr *LimitError

	if !errors.As(err, &limitErr) {
		t.Fatalf("expected ZIP buffer LimitError, got %v", err)
	}

	if limitErr.Size != int64(len(data)) || limitErr.Limit != int64(len(data)-1) {
		t.Fatalf("unexpected ZIP buffer limit error %#v", limitErr)
	}

	if filesystem.last == nil || !filesystem.last.closed {
		t.Fatal("expected rejected source to close")
	}
}

func zipBytesForTest(t *testing.T, name, body string) []byte {
	t.Helper()

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	entry, err := writer.Create(name)

	if err != nil {
		t.Fatalf("create ZIP entry: %v", err)
	}

	if _, err := entry.Write([]byte(body)); err != nil {
		t.Fatalf("write ZIP entry: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close ZIP: %v", err)
	}

	return buffer.Bytes()
}
