package core

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	ferretfs "github.com/MontFerret/ferret/v2/pkg/fs"
)

type sourceFactory struct {
	reader           ferretfs.Reader
	path             string
	maxZIPBufferSize int64
}

func newSourceFactory(ctx context.Context, source string, cfg Config) (*sourceFactory, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if strings.TrimSpace(source) == "" {
		return nil, fmt.Errorf("archive source must not be empty")
	}

	reader, err := ferretfs.ReaderFrom(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve archive filesystem reader: %w", err)
	}

	return &sourceFactory{
		reader:           reader,
		path:             source,
		maxZIPBufferSize: cfg.MaxZIPBufferSize,
	}, nil
}

func (f *sourceFactory) open(ctx context.Context) (ferretfs.ReadableFile, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	file, err := f.reader.Open(f.path)
	if err != nil {
		return nil, fmt.Errorf("open archive source %q: %w", f.path, err)
	}

	return file, nil
}

func (f *sourceFactory) resolveFormat(ctx context.Context, requested Format) (Format, error) {
	if requested != FormatAuto {
		return requested, nil
	}

	if detected := formatFromName(f.path); detected != FormatAuto {
		return detected, nil
	}

	return f.detectContent(ctx)
}

func (f *sourceFactory) detectContent(ctx context.Context) (_ Format, outErr error) {
	file, err := f.open(ctx)
	if err != nil {
		return "", err
	}

	defer func() {
		if err := file.Close(); err != nil {
			outErr = errors.Join(outErr, fmt.Errorf("close archive source %q after format detection: %w", f.path, err))
		}
	}()

	const probeSize = 1024

	probe, err := readUpTo(ctx, file, probeSize)
	if err != nil {
		return "", fmt.Errorf("probe archive source %q: %w", f.path, err)
	}

	if isZIPSignature(probe) {
		return FormatZIP, nil
	}

	if len(probe) >= 2 && probe[0] == 0x1f && probe[1] == 0x8b {
		isArchive, probeErr := probeGZIPTar(ctx, io.MultiReader(bytes.NewReader(probe), file))
		if probeErr != nil {
			return "", fmt.Errorf("probe gzip archive source %q: %w", f.path, probeErr)
		}

		if isArchive {
			return FormatTARGZ, nil
		}

		return "", fmt.Errorf("detect archive format for %q: gzip content is not a TAR archive", f.path)
	}

	if isTAR(probe) {
		return FormatTAR, nil
	}

	return "", fmt.Errorf("detect archive format for %q: unrecognized archive content", f.path)
}

func isZIPSignature(data []byte) bool {
	if len(data) < 4 || data[0] != 'P' || data[1] != 'K' {
		return false
	}

	return (data[2] == 3 && data[3] == 4) ||
		(data[2] == 5 && data[3] == 6) ||
		(data[2] == 7 && data[3] == 8)
}

func probeGZIPTar(ctx context.Context, source io.Reader) (_ bool, outErr error) {
	reader, err := gzip.NewReader(source)
	if err != nil {
		return false, err
	}

	defer func() {
		if err := reader.Close(); err != nil {
			outErr = errors.Join(outErr, fmt.Errorf("close gzip format probe: %w", err))
		}
	}()

	probe, err := readUpTo(ctx, reader, 513)
	if err != nil {
		return false, err
	}

	return isTAR(probe), nil
}

func isTAR(data []byte) bool {
	if len(data) < 512 {
		return false
	}

	if bytes.Equal(data[:512], make([]byte, 512)) {
		return true
	}

	_, err := tar.NewReader(bytes.NewReader(data)).Next()

	return err == nil
}

func readUpTo(ctx context.Context, reader io.Reader, limit int64) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	data, err := io.ReadAll(io.LimitReader(reader, limit))
	if err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return data, nil
}
