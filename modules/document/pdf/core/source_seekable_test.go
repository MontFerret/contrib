package core

import (
	"context"
	"io"
	"strings"
	"testing"

	ferretfs "github.com/MontFerret/ferret/v2/pkg/fs"
)

func TestSeekableFallbackAvoidsBuffering(t *testing.T) {
	t.Parallel()

	data := buildPDFForTest(t, []pdfTestPage{{Text: "Seekable"}})
	var opened *seekableMemoryFile
	filesystem := memoryFS{
		files: map[string][]byte{"seekable.pdf": data},
		newReadable: func(_ string, data []byte, info memoryFileInfo) ferretfs.ReadableFile {
			opened = newSeekableMemoryFile(data, info)

			return opened
		},
	}
	ctx := ferretfs.WithFileSystem(context.Background(), filesystem)

	document, err := Open(ctx, "seekable.pdf", OpenOptions{MaxBufferSize: 1})
	if err != nil {
		t.Fatalf("unexpected seekable open error: %v", err)
	}
	t.Cleanup(func() { _ = document.Close() })
	if opened == nil {
		t.Fatal("expected filesystem source to be opened")
	}
	if _, ok := any(opened).(io.ReaderAt); ok {
		t.Fatal("seekable test source unexpectedly implements io.ReaderAt")
	}
	if opened.closed {
		t.Fatal("expected seekable source to remain open")
	}
	if document.source.buffer != nil {
		t.Fatal("expected seekable source to avoid buffering")
	}

	text, err := document.Text(ctx)
	if err != nil {
		t.Fatalf("unexpected seekable text error: %v", err)
	}
	if !strings.Contains(text, "Seekable") {
		t.Fatalf("text %q does not contain fixture text", text)
	}

	if err := document.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
	if !opened.closed {
		t.Fatal("expected document close to close seekable source")
	}
}
