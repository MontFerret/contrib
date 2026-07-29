package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	stdfs "io/fs"
	"os"
	"path"

	ferretfs "github.com/MontFerret/ferret/v2/pkg/fs"
)

type destinationFS struct {
	filesystem         ferretfs.FileSystem
	createdDirectories []string
	createdFiles       []string
	createDirs         bool
	overwrite          bool
}

func newDestinationFS(
	filesystem ferretfs.FileSystem,
	createDirs bool,
	overwrite bool,
) *destinationFS {
	return &destinationFS{
		filesystem: filesystem,
		createDirs: createDirs,
		overwrite:  overwrite,
	}
}

func (d *destinationFS) preflightRoot(destination string) error {
	if destination == "" {
		return fmt.Errorf("archive destination must not be empty")
	}

	prefixes := sandboxPrefixes(destination)
	for index, prefix := range prefixes {
		info, exists, err := d.inspect(prefix)

		if err != nil {
			return err
		}

		if !exists {
			if !d.createDirs {
				return fmt.Errorf("destination directory %q does not exist", prefix)
			}

			return nil
		}

		if info.Mode()&stdfs.ModeSymlink != 0 {
			return fmt.Errorf("destination path %q is a symbolic link", prefix)
		}

		if !info.IsDir() {
			if index == len(prefixes)-1 {
				return fmt.Errorf("archive destination %q is not a directory", prefix)
			}

			return fmt.Errorf("destination parent %q is not a directory", prefix)
		}
	}

	return nil
}

func (d *destinationFS) preflightEntry(planned plannedEntry) error {
	if planned.skipLink {
		return nil
	}

	parent := planned.destination
	if planned.entry.kind == entryRegular {
		parent = path.Dir(planned.destination)
	}
	for _, prefix := range sandboxPrefixes(parent) {
		info, exists, err := d.inspect(prefix)
		if err != nil {
			return err
		}

		if !exists {
			if !d.createDirs {
				return fmt.Errorf("destination directory %q does not exist", prefix)
			}

			break
		}

		if info.Mode()&stdfs.ModeSymlink != 0 {
			return fmt.Errorf("destination path %q is a symbolic link", prefix)
		}

		if !info.IsDir() {
			return fmt.Errorf("destination parent %q is not a directory", prefix)
		}
	}

	info, exists, err := d.inspect(planned.destination)
	if err != nil {
		return err
	}

	if !exists {
		if planned.entry.kind == entryDirectory && !d.createDirs {
			return fmt.Errorf("destination directory %q does not exist", planned.destination)
		}

		return nil
	}

	if info.Mode()&stdfs.ModeSymlink != 0 {
		return fmt.Errorf("destination path %q is a symbolic link", planned.destination)
	}

	switch planned.entry.kind {
	case entryDirectory:
		if !info.IsDir() {
			return fmt.Errorf("cannot replace non-directory destination %q with a directory", planned.destination)
		}
	case entryRegular:
		if !info.Mode().IsRegular() {
			return fmt.Errorf("cannot replace non-regular destination %q with a file", planned.destination)
		}

		if !d.overwrite {
			return fmt.Errorf("destination file %q already exists", planned.destination)
		}
	}

	return nil
}

func (d *destinationFS) ensureDirectory(destination string, mode stdfs.FileMode) error {
	prefixes := sandboxPrefixes(destination)
	for index, prefix := range prefixes {
		info, exists, err := d.inspect(prefix)
		if err != nil {
			return err
		}

		if exists {
			if info.Mode()&stdfs.ModeSymlink != 0 {
				return fmt.Errorf("destination path %q is a symbolic link", prefix)
			}
			if !info.IsDir() {
				return fmt.Errorf("destination path %q is not a directory", prefix)
			}
			continue
		}

		if !d.createDirs {
			return fmt.Errorf("destination directory %q does not exist", prefix)
		}

		permission := stdfs.FileMode(0o777)
		if index == len(prefixes)-1 && mode.Perm() != 0 {
			permission = mode.Perm()
		}

		if err := d.filesystem.Mkdir(prefix, permission); err != nil {
			return fmt.Errorf("create destination directory %q: %w", prefix, err)
		}

		d.createdDirectories = append(d.createdDirectories, prefix)

		info, exists, err = d.inspect(prefix)
		if err != nil {
			return err
		}

		if !exists || info.Mode()&stdfs.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("destination directory %q failed link-safe verification", prefix)
		}
	}

	return nil
}

func (d *destinationFS) writeFile(
	ctx context.Context,
	planned plannedEntry,
	reader io.Reader,
	limit int64,
) (outErr error) {
	if err := d.ensureDirectory(path.Dir(planned.destination), 0o777); err != nil {
		return err
	}

	info, exists, err := d.inspect(planned.destination)
	if err != nil {
		return err
	}

	if exists {
		if info.Mode()&stdfs.ModeSymlink != 0 {
			return fmt.Errorf("destination path %q is a symbolic link", planned.destination)
		}

		if !info.Mode().IsRegular() {
			return fmt.Errorf("cannot replace non-regular destination %q with a file", planned.destination)
		}

		if !d.overwrite {
			return fmt.Errorf("destination file %q already exists", planned.destination)
		}

		if err := d.filesystem.Remove(planned.destination); err != nil {
			return fmt.Errorf("remove destination file %q for overwrite: %w", planned.destination, err)
		}
	}

	permission := planned.entry.fileMode.Perm()

	if permission == 0 {
		permission = 0o666
	}

	writer, err := d.filesystem.OpenFile(
		planned.destination,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		permission,
	)

	if err != nil {
		return fmt.Errorf("create destination file %q: %w", planned.destination, err)
	}

	created := true

	defer func() {
		if writer != nil {
			outErr = errors.Join(outErr, closeWithContext(writer, "close destination file", planned.destination))
		}

		if outErr != nil && created {
			if err := d.filesystem.Remove(planned.destination); err != nil && !errors.Is(err, stdfs.ErrNotExist) {
				outErr = errors.Join(outErr, fmt.Errorf("remove incomplete destination file %q: %w", planned.destination, err))
			}
		}
	}()

	if err := copyBounded(ctx, writer, reader, planned.entry.Name, limit); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		writer = nil
		return fmt.Errorf("close destination file %q: %w", planned.destination, err)
	}

	writer = nil
	created = false
	d.createdFiles = append(d.createdFiles, planned.destination)

	return nil
}

// rollback removes only paths created by this extraction, in dependency order.
func (d *destinationFS) rollback() error {
	var outErr error
	for index := len(d.createdFiles) - 1; index >= 0; index-- {
		target := d.createdFiles[index]

		if err := d.filesystem.Remove(target); err != nil && !errors.Is(err, stdfs.ErrNotExist) {
			outErr = errors.Join(outErr, fmt.Errorf("remove extracted file %q during rollback: %w", target, err))
		}
	}

	for index := len(d.createdDirectories) - 1; index >= 0; index-- {
		target := d.createdDirectories[index]

		if err := d.filesystem.Remove(target); err != nil && !errors.Is(err, stdfs.ErrNotExist) {
			outErr = errors.Join(outErr, fmt.Errorf("remove extracted directory %q during rollback: %w", target, err))
		}
	}

	d.createdFiles = nil
	d.createdDirectories = nil

	return outErr
}

func (d *destinationFS) inspect(name string) (stdfs.FileInfo, bool, error) {
	info, err := d.filesystem.Lstat(name)
	if err == nil {
		return info, true, nil
	}

	if errors.Is(err, stdfs.ErrNotExist) {
		return nil, false, nil
	}

	return nil, false, fmt.Errorf("inspect destination path %q: %w", name, err)
}
