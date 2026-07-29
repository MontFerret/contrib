package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type archiveTestEntry struct {
	name     string
	body     string
	linkName string
	mode     fs.FileMode
	typeflag byte
}

func writeZIPForTest(t *testing.T, root, name string, entries []archiveTestEntry) {
	t.Helper()

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, entry := range entries {
		header := &zip.FileHeader{
			Name:     entry.name,
			Method:   zip.Deflate,
			Modified: time.Date(2026, time.June, 9, 12, 30, 0, 0, time.UTC),
		}
		mode := entry.mode
		if mode == 0 {
			mode = 0o644
		}
		if entry.typeflag == tar.TypeDir {
			mode |= fs.ModeDir
			header.Name += "/"
		}
		if entry.typeflag == tar.TypeSymlink {
			mode |= fs.ModeSymlink
			entry.body = entry.linkName
		}
		header.SetMode(mode)

		target, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatalf("create ZIP entry %q: %v", entry.name, err)
		}
		if _, err := target.Write([]byte(entry.body)); err != nil {
			t.Fatalf("write ZIP entry %q: %v", entry.name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close ZIP fixture: %v", err)
	}

	writeFixture(t, root, name, buffer.Bytes())
}

func writeTARForTest(t *testing.T, root, name string, gzipEnabled bool, entries []archiveTestEntry) {
	t.Helper()

	var buffer bytes.Buffer
	var output = interface {
		Write([]byte) (int, error)
	}(&buffer)
	var gzipWriter *gzip.Writer
	if gzipEnabled {
		gzipWriter = gzip.NewWriter(&buffer)
		output = gzipWriter
	}

	writer := tar.NewWriter(output)
	for _, entry := range entries {
		typeflag := entry.typeflag
		if typeflag == 0 {
			typeflag = tar.TypeReg
		}
		mode := entry.mode.Perm()
		if mode == 0 {
			mode = 0o644
		}
		header := &tar.Header{
			Name:     entry.name,
			Mode:     int64(mode),
			Size:     int64(len(entry.body)),
			ModTime:  time.Date(2026, time.June, 9, 12, 30, 0, 0, time.UTC),
			Typeflag: typeflag,
			Linkname: entry.linkName,
		}
		if typeflag != tar.TypeReg {
			header.Size = 0
		}
		if err := writer.WriteHeader(header); err != nil {
			t.Fatalf("create TAR entry %q: %v", entry.name, err)
		}
		if header.Size > 0 {
			if _, err := writer.Write([]byte(entry.body)); err != nil {
				t.Fatalf("write TAR entry %q: %v", entry.name, err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close TAR fixture: %v", err)
	}
	if gzipWriter != nil {
		if err := gzipWriter.Close(); err != nil {
			t.Fatalf("close gzip fixture: %v", err)
		}
	}

	writeFixture(t, root, name, buffer.Bytes())
}

func writeFixture(t *testing.T, root, name string, data []byte) {
	t.Helper()

	target := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(target), 0o777); err != nil {
		t.Fatalf("create fixture directory: %v", err)
	}
	if err := os.WriteFile(target, data, 0o666); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}
