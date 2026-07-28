package core

import (
	"context"
	"errors"
	"fmt"
	"io"

	ferretfs "github.com/MontFerret/ferret/v2/pkg/fs"
)

// Extract safely writes eligible archive entries beneath destination.
func Extract(
	ctx context.Context,
	source string,
	destination string,
	config ExtractConfig,
) (_ []ExtractResult, outErr error) {
	if destination == "" {
		return nil, fmt.Errorf("archive destination must not be empty")
	}

	config, err := config.normalize()
	if err != nil {
		return nil, err
	}

	filter, err := newEntryFilter(config.Include, config.Exclude)
	if err != nil {
		return nil, err
	}

	filesystem, err := ferretfs.FileSystemFrom(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve extraction filesystem: %w", err)
	}

	factory, err := newSourceFactory(ctx, source, config.Config)
	if err != nil {
		return nil, err
	}

	format, err := factory.resolveFormat(ctx, config.Format)
	if err != nil {
		return nil, err
	}

	planned, err := buildManifest(ctx, factory, format, destination, config, config.Config, filter)
	if err != nil {
		return nil, err
	}

	if len(planned) == 0 {
		return []ExtractResult{}, nil
	}

	output := newDestinationFS(filesystem, config.CreateDirs, config.Overwrite)

	defer func() {
		if outErr != nil {
			outErr = errors.Join(outErr, output.rollback())
		}
	}()

	if err := output.preflightRoot(destination); err != nil {
		return nil, err
	}

	for _, item := range planned {
		if err := output.preflightEntry(item); err != nil {
			return nil, fmt.Errorf("preflight archive entry %q: %w", item.entry.Name, err)
		}
	}

	if err := output.ensureDirectory(destination, 0o777); err != nil {
		return nil, err
	}

	return executeExtraction(ctx, factory, format, planned, output, config.MaxEntrySize)
}

func buildManifest(
	ctx context.Context,
	factory *sourceFactory,
	format Format,
	destination string,
	opts ExtractConfig,
	cfg Config,
	filter *entryFilter,
) (_ []plannedEntry, outErr error) {
	builder := newManifestBuilder(filter, destination, opts.Links, cfg.MaxEntrySize)

	switch format {
	case FormatZIP:
		archive, err := factory.openZIP(ctx)
		if err != nil {
			return nil, err
		}

		defer func() {
			outErr = errors.Join(outErr, archive.close())
		}()

		for index, file := range archive.reader.File {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			entry, metadataErr := zipMetadata(ctx, file, format, false, cfg.MaxEntrySize)
			if metadataErr != nil {
				return nil, fmt.Errorf("inspect ZIP entry %q: %w", file.Name, metadataErr)
			}

			if err := builder.add(index, entry); err != nil {
				return nil, err
			}
		}

	case FormatTAR, FormatTARGZ:
		archive, err := factory.openTAR(ctx, format)
		if err != nil {
			return nil, err
		}

		defer func() {
			outErr = errors.Join(outErr, archive.close())
		}()

		for index := 0; ; index++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			header, nextErr := archive.reader.Next()
			if nextErr == io.EOF {
				break
			}

			if nextErr != nil {
				return nil, fmt.Errorf("preflight TAR source %q: %w", factory.path, nextErr)
			}

			entry, metadataErr := tarMetadata(header, format)
			if metadataErr != nil {
				return nil, fmt.Errorf("inspect TAR entry %q: %w", header.Name, metadataErr)
			}

			if err := builder.add(index, entry); err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("unsupported resolved archive format %q", format)
	}

	return builder.build()
}

func executeExtraction(
	ctx context.Context,
	factory *sourceFactory,
	format Format,
	planned []plannedEntry,
	output *destinationFS,
	limit int64,
) (_ []ExtractResult, outErr error) {
	byIndex := make(map[int]plannedEntry, len(planned))

	for _, item := range planned {
		byIndex[item.index] = item
	}

	results := make([]ExtractResult, 0, len(planned))

	switch format {
	case FormatZIP:
		archive, err := factory.openZIP(ctx)
		if err != nil {
			return nil, err
		}

		defer func() {
			outErr = errors.Join(outErr, archive.close())
		}()

		for index, file := range archive.reader.File {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			item, exists := byIndex[index]
			if !exists {
				continue
			}

			entry, metadataErr := zipMetadata(ctx, file, format, false, limit)
			if metadataErr != nil {
				return nil, metadataErr
			}

			if err := verifyPlannedEntry(item, entry); err != nil {
				return nil, err
			}

			result, extractErr := extractZIPEntry(ctx, output, item, file, limit)
			if extractErr != nil {
				return nil, extractErr
			}

			results = append(results, result)
		}

	case FormatTAR, FormatTARGZ:
		archive, err := factory.openTAR(ctx, format)
		if err != nil {
			return nil, err
		}

		defer func() {
			outErr = errors.Join(outErr, archive.close())
		}()

		for index := 0; ; index++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			header, nextErr := archive.reader.Next()
			if nextErr == io.EOF {
				break
			}

			if nextErr != nil {
				return nil, fmt.Errorf("extract TAR source %q: %w", factory.path, nextErr)
			}

			item, exists := byIndex[index]
			if !exists {
				continue
			}

			entry, metadataErr := tarMetadata(header, format)
			if metadataErr != nil {
				return nil, metadataErr
			}

			if err := verifyPlannedEntry(item, entry); err != nil {
				return nil, err
			}

			result, extractErr := extractEntry(ctx, output, item, archive.reader, limit)
			if extractErr != nil {
				return nil, extractErr
			}

			results = append(results, result)
		}
	}

	if len(results) != len(planned) {
		return nil, fmt.Errorf("archive source changed between extraction preflight and execution")
	}

	return results, nil
}

func extractZIPEntry(
	ctx context.Context,
	output *destinationFS,
	item plannedEntry,
	file interface{ Open() (io.ReadCloser, error) },
	limit int64,
) (ExtractResult, error) {
	if item.skipLink || item.entry.kind == entryDirectory {
		return extractEntry(ctx, output, item, nil, limit)
	}

	reader, err := file.Open()
	if err != nil {
		return ExtractResult{}, fmt.Errorf("open ZIP entry %q: %w", item.entry.Name, err)
	}

	result, extractErr := extractEntry(ctx, output, item, reader, limit)
	closeErr := closeWithContext(reader, "close ZIP entry", item.entry.Name)

	if extractErr != nil || closeErr != nil {
		return ExtractResult{}, errors.Join(extractErr, closeErr)
	}

	return result, nil
}

func extractEntry(
	ctx context.Context,
	output *destinationFS,
	item plannedEntry,
	reader io.Reader,
	limit int64,
) (ExtractResult, error) {
	if err := ctx.Err(); err != nil {
		return ExtractResult{}, err
	}

	if item.skipLink {
		reason := "link"

		return ExtractResult{
			Name:    item.canonical,
			Size:    item.entry.Size,
			IsDir:   false,
			Skipped: true,
			Reason:  &reason,
		}, nil
	}

	resultPath := item.destination
	result := ExtractResult{
		Name:  item.canonical,
		Path:  &resultPath,
		Size:  item.entry.Size,
		IsDir: item.entry.kind == entryDirectory,
	}

	switch item.entry.kind {
	case entryDirectory:
		if err := output.ensureDirectory(item.destination, item.entry.fileMode); err != nil {
			return ExtractResult{}, fmt.Errorf("extract directory %q: %w", item.entry.Name, err)
		}
	case entryRegular:
		if reader == nil {
			return ExtractResult{}, fmt.Errorf("extract file %q: entry reader is unavailable", item.entry.Name)
		}

		if err := output.writeFile(ctx, item, reader, limit); err != nil {
			return ExtractResult{}, fmt.Errorf("extract file %q: %w", item.entry.Name, err)
		}
	default:
		return ExtractResult{}, fmt.Errorf("extract entry %q: unsupported entry type", item.entry.Name)
	}

	return result, nil
}

func verifyPlannedEntry(planned plannedEntry, actual Entry) error {
	if planned.entry.Name != actual.Name ||
		planned.entry.Size != actual.Size ||
		planned.entry.kind != actual.kind {

		return fmt.Errorf("archive source changed between extraction preflight and execution at entry %q", actual.Name)
	}

	return nil
}
