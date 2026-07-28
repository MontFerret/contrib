package core

import (
	"context"
	"errors"
	"fmt"
	"io"
)

// List returns archive metadata without extracting entries.
func List(ctx context.Context, source string, config ListConfig) (_ []Entry, outErr error) {
	factory, err := newSourceFactory(ctx, source, config.Config)
	if err != nil {
		return nil, err
	}

	config, err = config.normalize()
	if err != nil {
		return nil, err
	}

	format, err := factory.resolveFormat(ctx, config.Format)
	if err != nil {
		return nil, err
	}

	switch format {
	case FormatZIP:
		archive, openErr := factory.openZIP(ctx)
		if openErr != nil {
			return nil, openErr
		}

		defer func() {
			outErr = errors.Join(outErr, archive.close())
		}()

		entries := make([]Entry, 0, len(archive.reader.File))
		for _, file := range archive.reader.File {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			entry, metadataErr := zipMetadata(ctx, file, format, true, config.MaxEntrySize)

			if metadataErr != nil {
				return nil, fmt.Errorf("list ZIP entry %q: %w", file.Name, metadataErr)
			}
			entries = append(entries, entry)
		}

		return entries, nil

	case FormatTAR, FormatTARGZ:
		archive, openErr := factory.openTAR(ctx, format)
		if openErr != nil {
			return nil, openErr
		}

		defer func() {
			outErr = errors.Join(outErr, archive.close())
		}()

		entries := make([]Entry, 0)
		for {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			header, nextErr := archive.reader.Next()
			if nextErr == io.EOF {
				break
			}

			if nextErr != nil {
				return nil, fmt.Errorf("list TAR source %q: %w", source, nextErr)
			}

			entry, metadataErr := tarMetadata(header, format)
			if metadataErr != nil {
				return nil, fmt.Errorf("list TAR entry %q: %w", header.Name, metadataErr)
			}

			entries = append(entries, entry)
		}

		return entries, nil
	default:
		return nil, fmt.Errorf("unsupported resolved archive format %q", format)
	}
}
