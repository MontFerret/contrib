package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	stdfs "io/fs"
	"strings"

	ferretfs "github.com/MontFerret/ferret/v2/pkg/fs"
)

type pdfSource struct {
	reader io.ReaderAt
	closer io.Closer
	buffer []byte
	size   int64
}

func openSource(ctx context.Context, path string, opts OpenOptions) (*pdfSource, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("path must not be empty")
	}

	reader, err := ferretfs.ReaderFrom(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve filesystem: %w", err)
	}

	info, err := reader.Stat(path)
	if err != nil {
		if errors.Is(err, stdfs.ErrNotExist) {
			return nil, fmt.Errorf("failed to open PDF document %q: file does not exist", path)
		}

		return nil, fmt.Errorf("failed to inspect PDF document %q: %w", path, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("failed to open PDF document %q: path is a directory", path)
	}
	if info.Size() < 0 {
		return nil, fmt.Errorf("failed to open PDF document %q: invalid file size %d", path, info.Size())
	}

	file, err := reader.Open(path)
	if err != nil {
		if errors.Is(err, stdfs.ErrNotExist) {
			return nil, fmt.Errorf("failed to open PDF document %q: file does not exist", path)
		}

		return nil, fmt.Errorf("failed to open PDF document %q: %w", path, err)
	}

	if readerAt, ok := file.(io.ReaderAt); ok {
		return &pdfSource{
			reader: readerAt,
			closer: file,
			size:   info.Size(),
		}, nil
	}

	data, err := bufferSource(ctx, file, path, info.Size(), opts.MaxBufferSize)
	if err != nil {
		return nil, err
	}

	return &pdfSource{
		reader: bytes.NewReader(data),
		buffer: data,
		size:   int64(len(data)),
	}, nil
}

func (s *pdfSource) close() error {
	if s == nil {
		return nil
	}

	s.reader = nil
	s.buffer = nil

	if s.closer == nil {
		return nil
	}

	err := s.closer.Close()
	s.closer = nil

	return err
}
