package core

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
)

type seekableMemoryFile struct {
	reader *bytes.Reader
	info   memoryFileInfo
	closed bool
}

func newSeekableMemoryFile(data []byte, info memoryFileInfo) *seekableMemoryFile {
	return &seekableMemoryFile{
		reader: bytes.NewReader(data),
		info:   info,
	}
}

func (f *seekableMemoryFile) Stat() (fs.FileInfo, error) {
	return f.info, nil
}

func (f *seekableMemoryFile) Read(data []byte) (int, error) {
	if f.closed {
		return 0, errors.New("file is closed")
	}

	return f.reader.Read(data)
}

func (f *seekableMemoryFile) Seek(offset int64, whence int) (int64, error) {
	if f.closed {
		return 0, errors.New("file is closed")
	}

	return f.reader.Seek(offset, whence)
}

func (f *seekableMemoryFile) Close() error {
	f.closed = true

	return nil
}

var (
	_ io.ReadSeeker = (*seekableMemoryFile)(nil)
	_ io.Closer     = (*seekableMemoryFile)(nil)
)
