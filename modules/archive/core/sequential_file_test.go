package core

import (
	"bytes"
	"io/fs"
)

type sequentialFile struct {
	reader *bytes.Reader
	info   sequentialFileInfo
	closed bool
}

func (f *sequentialFile) Read(p []byte) (int, error) {
	return f.reader.Read(p)
}

func (f *sequentialFile) Close() error {
	f.closed = true

	return nil
}

func (f *sequentialFile) Stat() (fs.FileInfo, error) {
	return f.info, nil
}
