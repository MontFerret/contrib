package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/MontFerret/barn/pkg/publish"
)

const moduleTagPrefix = "modules/"

var moduleSegmentPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9._-]*[a-z0-9])?$`)

type (
	options struct {
		sourceRoot string
		barnRoot   string
		tag        string
	}

	releaseTarget struct {
		module  string
		version string
	}

	prepareFunc func(context.Context, publish.Request) (*publish.Result, error)
)

func parseOptions(arguments []string, output io.Writer) (options, error) {
	var result options
	flags := flag.NewFlagSet("publish", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.StringVar(&result.sourceRoot, "source-root", ".", "path to the contrib repository root")
	flags.StringVar(&result.barnRoot, "barn-root", "", "path to a Barn repository checkout")
	flags.StringVar(&result.tag, "tag", "", "module release tag in modules/<module>/v<version> form")

	if err := flags.Parse(arguments); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 {
		return options{}, fmt.Errorf("unexpected positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	if result.barnRoot == "" {
		return options{}, errors.New("--barn-root is required")
	}
	if result.tag == "" {
		return options{}, errors.New("--tag is required")
	}

	return result, nil
}

func execute(ctx context.Context, options options, prepare prepareFunc, output io.Writer) error {
	target, err := parseModuleTag(options.tag)
	if err != nil {
		return err
	}

	moduleDirectory := filepath.Join(
		options.sourceRoot,
		"modules",
		filepath.FromSlash(target.module),
	)

	result, err := prepare(ctx, publish.Request{
		Directory: moduleDirectory,
		Tag:       options.tag,
	})
	if err != nil {
		if errors.Is(err, publish.ErrVersionAlreadyPublished) {
			fmt.Fprintf(output, "%s v%s is already published; no Barn records changed.\n", target.module, target.version)
			return nil
		}

		return fmt.Errorf("prepare Barn records for %s: %w", options.tag, err)
	}
	if result == nil {
		return errors.New("barn publisher returned no result")
	}
	if result.Version == nil {
		return errors.New("barn publisher returned no version record")
	}
	if result.Version.Version != target.version {
		return fmt.Errorf(
			"prepared registry version %q does not match tag version %q",
			result.Version.Version,
			target.version,
		)
	}
	if result.Version.Tag != options.tag {
		return fmt.Errorf(
			"prepared registry tag %q does not match release tag %q",
			result.Version.Tag,
			options.tag,
		)
	}
	if len(result.Files) == 0 {
		return errors.New("barn publisher returned no registry records")
	}

	created, unchanged, err := materializeFiles(options.barnRoot, result.Files)
	if err != nil {
		return fmt.Errorf("materialize Barn records: %w", err)
	}

	fmt.Fprintf(
		output,
		"Prepared %s publication for %s v%s: %d created, %d unchanged.\n",
		result.Kind,
		target.module,
		target.version,
		created,
		unchanged,
	)

	return nil
}

func parseModuleTag(tag string) (releaseTarget, error) {
	if !strings.HasPrefix(tag, moduleTagPrefix) {
		return releaseTarget{}, fmt.Errorf("tag %q must start with %q", tag, moduleTagPrefix)
	}

	segments := strings.Split(strings.TrimPrefix(tag, moduleTagPrefix), "/")
	if len(segments) < 2 {
		return releaseTarget{}, fmt.Errorf("tag %q must match modules/<module>/v<version>", tag)
	}

	versionSegment := segments[len(segments)-1]
	if !strings.HasPrefix(versionSegment, "v") || len(versionSegment) == 1 {
		return releaseTarget{}, fmt.Errorf("tag %q must end with v<version>", tag)
	}

	version := strings.TrimPrefix(versionSegment, "v")
	if _, err := semver.StrictNewVersion(version); err != nil {
		return releaseTarget{}, fmt.Errorf("tag %q has invalid semantic version: %w", tag, err)
	}

	moduleSegments := segments[:len(segments)-1]
	for _, segment := range moduleSegments {
		if !moduleSegmentPattern.MatchString(segment) {
			return releaseTarget{}, fmt.Errorf("tag %q has invalid module path segment %q", tag, segment)
		}
	}

	return releaseTarget{
		module:  strings.Join(moduleSegments, "/"),
		version: version,
	}, nil
}

func materializeFiles(rootPath string, files []publish.File) (int, int, error) {
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return 0, 0, fmt.Errorf("open Barn root: %w", err)
	}
	defer root.Close()

	created := 0
	unchanged := 0
	seen := make(map[string]struct{}, len(files))

	for _, file := range files {
		relativePath, err := validateRecordPath(file.Path)
		if err != nil {
			return created, unchanged, err
		}
		if _, exists := seen[relativePath]; exists {
			return created, unchanged, fmt.Errorf("barn returned duplicate record path %q", file.Path)
		}
		seen[relativePath] = struct{}{}

		existing, err := root.ReadFile(relativePath)
		switch {
		case err == nil:
			if !bytes.Equal(existing, file.Content) {
				return created, unchanged, fmt.Errorf("record %q already exists with different content", file.Path)
			}
			unchanged++
			continue
		case !errors.Is(err, os.ErrNotExist):
			return created, unchanged, fmt.Errorf("read existing record %q: %w", file.Path, err)
		}

		parent := filepath.Dir(relativePath)
		if parent != "." {
			if err := root.MkdirAll(parent, 0o755); err != nil {
				return created, unchanged, fmt.Errorf("create record directory for %q: %w", file.Path, err)
			}
		}

		if err := writeNewFile(root, relativePath, file.Content); err != nil {
			return created, unchanged, fmt.Errorf("write record %q: %w", file.Path, err)
		}
		created++
	}

	return created, unchanged, nil
}

func validateRecordPath(value string) (string, error) {
	if value == "" || path.IsAbs(value) || strings.Contains(value, `\`) {
		return "", fmt.Errorf("barn returned invalid record path %q", value)
	}
	if cleaned := path.Clean(value); cleaned != value || cleaned == "." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("barn returned invalid record path %q", value)
	}
	if !strings.HasPrefix(value, "registry/modules/") {
		return "", fmt.Errorf("barn returned record path outside registry/modules: %q", value)
	}

	return filepath.FromSlash(value), nil
}

func writeNewFile(root *os.Root, name string, content []byte) error {
	file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}

	_, writeErr := file.Write(content)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}

	return closeErr
}
