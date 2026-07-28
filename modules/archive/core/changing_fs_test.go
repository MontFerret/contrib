package core

import (
	"bytes"
	"sync"

	ferretfs "github.com/MontFerret/ferret/v2/pkg/fs"
)

type changingFS struct {
	ferretfs.FileSystem
	preflight []byte
	execution []byte
	mu        sync.Mutex
	opens     int
}

func newChangingFS(
	filesystem ferretfs.FileSystem,
	preflight []byte,
	execution []byte,
) *changingFS {
	return &changingFS{
		FileSystem: filesystem,
		preflight:  preflight,
		execution:  execution,
	}
}

func (f *changingFS) Open(name string) (ferretfs.ReadableFile, error) {
	if name != "archive.tar" {
		return f.FileSystem.Open(name)
	}

	f.mu.Lock()
	data := f.execution

	if f.opens == 0 {
		data = f.preflight
	}

	f.opens++
	f.mu.Unlock()

	return &sequentialFile{
		reader: bytes.NewReader(data),
		info:   sequentialFileInfo{size: int64(len(data))},
	}, nil
}
