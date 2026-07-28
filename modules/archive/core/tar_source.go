package core

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
)

type tarSource struct {
	reader *tar.Reader
	gzip   *gzip.Reader
	closer io.Closer
	path   string
}

func (f *sourceFactory) openTAR(ctx context.Context, format Format) (*tarSource, error) {
	file, err := f.open(ctx)
	if err != nil {
		return nil, err
	}

	source := &tarSource{
		closer: file,
		path:   f.path,
	}

	var reader io.Reader = file
	if format == FormatTARGZ {
		gzipReader, gzipErr := gzip.NewReader(file)
		if gzipErr != nil {
			return nil, errors.Join(
				fmt.Errorf("decompress TAR.GZ source %q: %w", f.path, gzipErr),
				source.close(),
			)
		}

		source.gzip = gzipReader
		reader = gzipReader
	}

	source.reader = tar.NewReader(reader)

	return source, nil
}

func (s *tarSource) close() error {
	if s == nil {
		return nil
	}

	var outErr error

	if s.gzip != nil {
		if err := s.gzip.Close(); err != nil {
			outErr = fmt.Errorf("close gzip reader for %q: %w", s.path, err)
		}

		s.gzip = nil
	}

	if err := closeWithContext(s.closer, "close TAR source", s.path); err != nil {
		outErr = errors.Join(outErr, err)
	}

	s.reader = nil
	s.closer = nil

	return outErr
}
