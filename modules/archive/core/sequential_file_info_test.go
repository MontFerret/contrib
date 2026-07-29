package core

import (
	"io/fs"
	"time"
)

type sequentialFileInfo struct {
	size int64
}

func (i sequentialFileInfo) Name() string       { return "archive.zip" }
func (i sequentialFileInfo) Size() int64        { return i.size }
func (i sequentialFileInfo) Mode() fs.FileMode  { return 0o600 }
func (i sequentialFileInfo) ModTime() time.Time { return time.Time{} }
func (i sequentialFileInfo) IsDir() bool        { return false }
func (i sequentialFileInfo) Sys() any           { return nil }
