package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	stdfs "io/fs"
	"path"

	"github.com/xuri/excelize/v2"

	ferretfs "github.com/MontFerret/ferret/v2/pkg/fs"
)

func openWorkbookFile(ctx context.Context, workbookPath string) (*excelize.File, error) {
	reader, err := ferretfs.ReaderFrom(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve filesystem: %w", err)
	}

	data, err := reader.ReadFile(workbookPath)
	if err != nil {
		if errors.Is(err, stdfs.ErrNotExist) {
			return nil, fmt.Errorf("failed to open XLSX workbook %q: file does not exist", workbookPath)
		}

		return nil, fmt.Errorf("failed to open XLSX workbook %q: %w", workbookPath, err)
	}

	file, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to open XLSX workbook %q: %w", workbookPath, err)
	}

	return file, nil
}

func ensureWorkbookParentDirectory(ctx context.Context, workbookPath string) error {
	parent := path.Dir(path.Clean(workbookPath))
	if parent == "." {
		return nil
	}

	reader, err := ferretfs.ReaderFrom(ctx)
	if err != nil {
		return fmt.Errorf("resolve filesystem: %w", err)
	}

	info, err := reader.Stat(parent)
	if err != nil {
		if errors.Is(err, stdfs.ErrNotExist) {
			return fmt.Errorf("parent directory %q does not exist", parent)
		}

		return fmt.Errorf("inspect parent directory %q: %w", parent, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("parent path %q is not a directory", parent)
	}

	return nil
}

func writeWorkbookFile(ctx context.Context, file *excelize.File, workbookPath string) error {
	writer, err := ferretfs.WriterFrom(ctx)
	if err != nil {
		return fmt.Errorf("resolve filesystem: %w", err)
	}

	data, err := file.WriteToBuffer()
	if err != nil {
		return fmt.Errorf("serialize XLSX workbook %q: %w", workbookPath, err)
	}

	if err := writer.WriteFile(workbookPath, data.Bytes(), 0666); err != nil {
		return fmt.Errorf("write XLSX workbook %q: %w", workbookPath, err)
	}

	return nil
}
