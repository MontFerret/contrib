package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Read returns the first regular entry matching name. The boolean reports
// whether a matching regular entry was found.
func Read(
	ctx context.Context,
	source string,
	name string,
	config ReadConfig,
) (_ []byte, found bool, outErr error) {
	if strings.TrimSpace(name) == "" {
		return nil, false, fmt.Errorf("archive entry name must not be empty")
	}

	config, err := config.normalize()
	if err != nil {
		return nil, false, err
	}

	factory, err := newSourceFactory(ctx, source, config.Config)
	if err != nil {
		return nil, false, err
	}

	format, err := factory.resolveFormat(ctx, config.Format)
	if err != nil {
		return nil, false, err
	}

	switch format {
	case FormatZIP:
		archive, openErr := factory.openZIP(ctx)
		if openErr != nil {
			return nil, false, openErr
		}

		defer func() {
			outErr = errors.Join(outErr, archive.close())
		}()

		matchedNonRegular := false
		for _, file := range archive.reader.File {
			if err := ctx.Err(); err != nil {
				return nil, false, err
			}

			if file.Name != name {
				continue
			}

			entry, metadataErr := zipMetadata(ctx, file, format, false, config.MaxEntrySize)
			if metadataErr != nil {
				return nil, false, fmt.Errorf("inspect ZIP entry %q: %w", file.Name, metadataErr)
			}

			if entry.kind != entryRegular {
				matchedNonRegular = true
				continue
			}

			if entry.Size > config.MaxEntrySize {
				return nil, false, &LimitError{Entry: entry.Name, Size: entry.Size, Limit: config.MaxEntrySize}
			}

			reader, openEntryErr := file.Open()
			if openEntryErr != nil {
				return nil, false, fmt.Errorf("open ZIP entry %q: %w", file.Name, openEntryErr)
			}

			data, readErr := readBounded(ctx, reader, file.Name, config.MaxEntrySize)
			closeErr := closeWithContext(reader, "close ZIP entry", file.Name)

			if readErr != nil || closeErr != nil {
				return nil, false, errors.Join(readErr, closeErr)
			}

			return data, true, nil
		}

		if matchedNonRegular {
			return nil, false, fmt.Errorf("archive entry %q is not a regular file", name)
		}

		return nil, false, nil

	case FormatTAR, FormatTARGZ:
		archive, openErr := factory.openTAR(ctx, format)
		if openErr != nil {
			return nil, false, openErr
		}

		defer func() {
			outErr = errors.Join(outErr, archive.close())
		}()

		matchedNonRegular := false

		for {
			if err := ctx.Err(); err != nil {
				return nil, false, err
			}

			header, nextErr := archive.reader.Next()
			if nextErr == io.EOF {
				break
			}

			if nextErr != nil {
				return nil, false, fmt.Errorf("read TAR source %q: %w", source, nextErr)
			}

			if header.Name != name {
				continue
			}

			entry, metadataErr := tarMetadata(header, format)
			if metadataErr != nil {
				return nil, false, fmt.Errorf("inspect TAR entry %q: %w", header.Name, metadataErr)
			}

			if entry.kind != entryRegular {
				matchedNonRegular = true
				continue
			}

			if entry.Size > config.MaxEntrySize {
				return nil, false, &LimitError{Entry: entry.Name, Size: entry.Size, Limit: config.MaxEntrySize}
			}

			data, readErr := readBounded(ctx, archive.reader, header.Name, config.MaxEntrySize)
			if readErr != nil {
				return nil, false, readErr
			}

			return data, true, nil
		}

		if matchedNonRegular {
			return nil, false, fmt.Errorf("archive entry %q is not a regular file", name)
		}

		return nil, false, nil
	default:
		return nil, false, fmt.Errorf("unsupported resolved archive format %q", format)
	}
}
