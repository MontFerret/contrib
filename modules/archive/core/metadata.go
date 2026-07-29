package core

import (
	"archive/tar"
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"math"
)

func zipMetadata(
	ctx context.Context,
	file *zip.File,
	format Format,
	readLinkTarget bool,
	maxEntrySize int64,
) (Entry, error) {
	if file.UncompressedSize64 > math.MaxInt64 {
		return Entry{}, fmt.Errorf("ZIP entry %q has an unsupported size", file.Name)
	}

	if file.CompressedSize64 > math.MaxInt64 {
		return Entry{}, fmt.Errorf("ZIP entry %q has an unsupported compressed size", file.Name)
	}

	mode := file.Mode()
	kind := classifyZIP(mode, file.FileInfo().IsDir())
	compressedSize := int64(file.CompressedSize64)
	linkName := ""

	if kind == entrySymlink && readLinkTarget {
		if int64(file.UncompressedSize64) > maxEntrySize {
			return Entry{}, &LimitError{
				Entry: file.Name,
				Size:  int64(file.UncompressedSize64),
				Limit: maxEntrySize,
			}
		}

		reader, err := file.Open()
		if err != nil {
			return Entry{}, fmt.Errorf("open ZIP symlink entry %q: %w", file.Name, err)
		}

		data, readErr := readBounded(ctx, reader, file.Name, maxEntrySize)
		closeErr := closeWithContext(reader, "close ZIP symlink entry", file.Name)
		if readErr != nil || closeErr != nil {
			return Entry{}, errors.Join(readErr, closeErr)
		}

		linkName = string(data)
	}

	return newEntry(
		file.Name,
		int64(file.UncompressedSize64),
		&compressedSize,
		mode,
		file.Modified,
		linkName,
		format,
		kind,
	), nil
}

func tarMetadata(header *tar.Header, format Format) (Entry, error) {
	if header.Size < 0 {
		return Entry{}, fmt.Errorf("TAR entry %q has a negative size %d", header.Name, header.Size)
	}

	kind := classifyTAR(header.Typeflag)

	return newEntry(
		header.Name,
		header.Size,
		nil,
		header.FileInfo().Mode(),
		header.ModTime,
		header.Linkname,
		format,
		kind,
	), nil
}

func classifyZIP(mode fs.FileMode, isDir bool) entryKind {
	switch {
	case isDir:
		return entryDirectory
	case mode&fs.ModeSymlink != 0:
		return entrySymlink
	case mode.IsRegular():
		return entryRegular
	default:
		return entrySpecial
	}
}

func classifyTAR(typeflag byte) entryKind {
	switch typeflag {
	case tar.TypeReg, 0:
		return entryRegular
	case tar.TypeDir:
		return entryDirectory
	case tar.TypeSymlink:
		return entrySymlink
	case tar.TypeLink:
		return entryHardlink
	default:
		return entrySpecial
	}
}
