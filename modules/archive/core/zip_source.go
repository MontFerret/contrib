package core

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"

	commonstream "github.com/MontFerret/contrib/pkg/common/stream"
	ferretfs "github.com/MontFerret/ferret/v2/pkg/fs"
)

type zipSource struct {
	closer io.Closer
	reader *zip.Reader
	path   string
	buffer []byte
}

func (f *sourceFactory) openZIP(ctx context.Context) (*zipSource, error) {
	file, err := f.open(ctx)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("inspect ZIP source %q: %w", f.path, err),
			closeWithContext(file, "close ZIP source", f.path),
		)
	}

	if info.IsDir() {
		return nil, errors.Join(
			fmt.Errorf("open ZIP source %q: path is a directory", f.path),
			closeWithContext(file, "close ZIP source", f.path),
		)
	}

	if info.Size() < 0 {
		return nil, errors.Join(
			fmt.Errorf("open ZIP source %q: invalid size %d", f.path, info.Size()),
			closeWithContext(file, "close ZIP source", f.path),
		)
	}

	if readerAt, ok := file.(io.ReaderAt); ok {
		return newZIPSource(readerAt, info.Size(), file, nil, f.path)
	}

	if seeker, ok := file.(io.ReadSeeker); ok {
		return newZIPSource(commonstream.NewReaderAt(seeker), info.Size(), file, nil, f.path)
	}

	data, readErr := bufferZIPSource(ctx, file, f.path, info.Size(), f.maxZIPBufferSize)
	if readErr != nil {
		return nil, readErr
	}

	return newZIPSource(bytes.NewReader(data), int64(len(data)), nil, data, f.path)
}

func newZIPSource(reader io.ReaderAt, size int64, closer io.Closer, buffer []byte, path string) (*zipSource, error) {
	archive, err := zip.NewReader(reader, size)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("parse ZIP source %q: %w", path, err),
			closeWithContext(closer, "close ZIP source", path),
		)
	}

	return &zipSource{
		reader: archive,
		closer: closer,
		buffer: buffer,
		path:   path,
	}, nil
}

func bufferZIPSource(
	ctx context.Context,
	file ferretfs.ReadableFile,
	path string,
	declaredSize int64,
	limit int64,
) (_ []byte, outErr error) {
	defer func() {
		if err := file.Close(); err != nil {
			outErr = errors.Join(outErr, fmt.Errorf("close ZIP source %q: %w", path, err))
		}
	}()

	if declaredSize > limit {
		return nil, &LimitError{Entry: path, Size: declaredSize, Limit: limit}
	}

	data, err := commonstream.ReadAllLimited(ctx, file, limit)
	if err != nil {
		if errors.Is(err, commonstream.ErrLimitExceeded) {
			observedSize := int64(-1)

			if limit < math.MaxInt64 {
				observedSize = limit + 1
			}

			return nil, &LimitError{Entry: path, Size: observedSize, Limit: limit}
		}

		return nil, fmt.Errorf("buffer ZIP source %q: %w", path, err)
	}

	return data, nil
}

func (s *zipSource) close() error {
	if s == nil {
		return nil
	}

	s.reader = nil
	s.buffer = nil
	err := closeWithContext(s.closer, "close ZIP source", s.path)
	s.closer = nil

	return err
}

func closeWithContext(closer io.Closer, operation, name string) error {
	if closer == nil {
		return nil
	}

	if err := closer.Close(); err != nil {
		return fmt.Errorf("%s %q: %w", operation, name, err)
	}

	return nil
}
