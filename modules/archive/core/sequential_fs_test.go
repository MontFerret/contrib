package core

import (
	"bytes"
	"fmt"

	ferretfs "github.com/MontFerret/ferret/v2/pkg/fs"
)

type sequentialFS struct {
	ferretfs.FileSystem
	last *sequentialFile
	data []byte
}

func (f *sequentialFS) Open(name string) (ferretfs.ReadableFile, error) {
	if name != "archive.zip" {
		return nil, fmt.Errorf("unexpected path %q", name)
	}

	f.last = &sequentialFile{
		reader: bytes.NewReader(f.data),
		info:   sequentialFileInfo{size: int64(len(f.data))},
	}

	return f.last, nil
}
