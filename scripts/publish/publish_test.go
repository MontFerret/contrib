package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MontFerret/barn/pkg/publish"
	registryspec "github.com/MontFerret/specs/pkg/registry"
)

func TestParseModuleTag(t *testing.T) {
	t.Parallel()

	tests := map[string]releaseTarget{
		"modules/xml/v1.2.3":                       {module: "xml", version: "1.2.3"},
		"modules/security/jwt/v1.0.0-rc.13":        {module: "security/jwt", version: "1.0.0-rc.13"},
		"modules/document/pdf/v2.0.0-beta.1+build": {module: "document/pdf", version: "2.0.0-beta.1+build"},
	}

	for tag, expected := range tests {
		t.Run(tag, func(t *testing.T) {
			t.Parallel()

			actual, err := parseModuleTag(tag)
			if err != nil {
				t.Fatalf("parse tag: %v", err)
			}
			if actual != expected {
				t.Fatalf("target = %#v, want %#v", actual, expected)
			}
		})
	}
}

func TestParseModuleTagRejectsInvalidTags(t *testing.T) {
	t.Parallel()

	tags := []string{
		"pkg/common/v1.0.0",
		"modules/xml",
		"modules/xml/1.0.0",
		"modules/xml/v",
		"modules/xml/v1.0",
		"modules//v1.0.0",
		"modules/../v1.0.0",
		"modules/security/jwt/v1.0.0/extra",
		`modules/security\jwt/v1.0.0`,
	}

	for _, tag := range tags {
		t.Run(tag, func(t *testing.T) {
			t.Parallel()

			if _, err := parseModuleTag(tag); err == nil {
				t.Fatal("expected tag to be rejected")
			}
		})
	}
}

func TestExecutePreparesNewModule(t *testing.T) {
	t.Parallel()

	sourceRoot := t.TempDir()
	barnRoot := t.TempDir()
	var request publish.Request
	output := &strings.Builder{}

	err := execute(context.Background(), options{
		sourceRoot: sourceRoot,
		barnRoot:   barnRoot,
		tag:        "modules/security/jwt/v1.0.0-rc.13",
	}, func(_ context.Context, actual publish.Request) (*publish.Result, error) {
		request = actual
		return &publish.Result{
			Kind: publish.NewModule,
			Version: &registryspec.VersionRecord{
				Version: "1.0.0-rc.13",
				Tag:     "modules/security/jwt/v1.0.0-rc.13",
			},
			Files: []publish.File{
				{Path: "registry/modules/montferret/jwt/manifest.json", Content: []byte("manifest\n")},
				{Path: "registry/modules/montferret/jwt/versions/v1.0.0-rc.13.json", Content: []byte("version\n")},
			},
		}, nil
	}, output)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	expectedDirectory := filepath.Join(sourceRoot, "modules", "security", "jwt")
	if request.Directory != expectedDirectory {
		t.Errorf("request directory = %q, want %q", request.Directory, expectedDirectory)
	}
	if request.Tag != "modules/security/jwt/v1.0.0-rc.13" {
		t.Errorf("request tag = %q", request.Tag)
	}

	assertFileContent(t, barnRoot, "registry/modules/montferret/jwt/manifest.json", "manifest\n")
	assertFileContent(t, barnRoot, "registry/modules/montferret/jwt/versions/v1.0.0-rc.13.json", "version\n")
	if !strings.Contains(output.String(), "2 created, 0 unchanged") {
		t.Errorf("output = %q", output.String())
	}
}

func TestExecutePreparesNewVersion(t *testing.T) {
	t.Parallel()

	barnRoot := t.TempDir()
	output := &strings.Builder{}
	err := execute(context.Background(), options{
		sourceRoot: t.TempDir(),
		barnRoot:   barnRoot,
		tag:        "modules/xml/v1.2.3",
	}, func(context.Context, publish.Request) (*publish.Result, error) {
		return &publish.Result{
			Kind: publish.NewVersion,
			Version: &registryspec.VersionRecord{
				Version: "1.2.3",
				Tag:     "modules/xml/v1.2.3",
			},
			Files: []publish.File{{
				Path:    "registry/modules/montferret/xml/versions/v1.2.3.json",
				Content: []byte("version\n"),
			}},
		}, nil
	}, output)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	assertFileContent(t, barnRoot, "registry/modules/montferret/xml/versions/v1.2.3.json", "version\n")
	if !strings.Contains(output.String(), "new-version") {
		t.Errorf("output = %q", output.String())
	}
}

func TestExecuteRejectsVersionThatDoesNotMatchTag(t *testing.T) {
	t.Parallel()

	barnRoot := t.TempDir()
	err := execute(context.Background(), options{
		sourceRoot: t.TempDir(),
		barnRoot:   barnRoot,
		tag:        "modules/xml/v1.2.3",
	}, func(context.Context, publish.Request) (*publish.Result, error) {
		return &publish.Result{
			Kind: publish.NewVersion,
			Version: &registryspec.VersionRecord{
				Version: "1.2.4",
				Tag:     "modules/xml/v1.2.3",
			},
			Files: []publish.File{{
				Path:    "registry/modules/montferret/xml/versions/v1.2.4.json",
				Content: []byte("version\n"),
			}},
		}, nil
	}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "does not match tag version") {
		t.Fatalf("error = %v, want version-mismatch error", err)
	}
}

func TestExecuteTreatsPublishedVersionAsNoOp(t *testing.T) {
	t.Parallel()

	output := &strings.Builder{}
	err := execute(context.Background(), options{
		sourceRoot: t.TempDir(),
		barnRoot:   filepath.Join(t.TempDir(), "missing"),
		tag:        "modules/xml/v1.2.3",
	}, func(context.Context, publish.Request) (*publish.Result, error) {
		return nil, errors.Join(errors.New("registry rejected version"), publish.ErrVersionAlreadyPublished)
	}, output)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(output.String(), "already published") {
		t.Errorf("output = %q", output.String())
	}
}

func TestMaterializeFilesAllowsIdenticalRecords(t *testing.T) {
	t.Parallel()

	barnRoot := t.TempDir()
	relativePath := "registry/modules/montferret/xml/versions/v1.2.3.json"
	absolutePath := filepath.Join(barnRoot, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatalf("create parent: %v", err)
	}
	if err := os.WriteFile(absolutePath, []byte("same\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	created, unchanged, err := materializeFiles(barnRoot, []publish.File{{
		Path:    relativePath,
		Content: []byte("same\n"),
	}})
	if err != nil {
		t.Fatalf("materialize: %v", err)
	}
	if created != 0 || unchanged != 1 {
		t.Fatalf("created, unchanged = %d, %d; want 0, 1", created, unchanged)
	}
}

func TestMaterializeFilesRejectsConflictingRecord(t *testing.T) {
	t.Parallel()

	barnRoot := t.TempDir()
	relativePath := "registry/modules/montferret/xml/versions/v1.2.3.json"
	absolutePath := filepath.Join(barnRoot, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatalf("create parent: %v", err)
	}
	if err := os.WriteFile(absolutePath, []byte("original\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	_, _, err := materializeFiles(barnRoot, []publish.File{{
		Path:    relativePath,
		Content: []byte("changed\n"),
	}})
	if err == nil || !strings.Contains(err.Error(), "different content") {
		t.Fatalf("error = %v, want conflicting-content error", err)
	}
	assertFileContent(t, barnRoot, relativePath, "original\n")
}

func TestMaterializeFilesRejectsUnsafePaths(t *testing.T) {
	t.Parallel()

	paths := []string{
		"../outside.json",
		"/registry/modules/outside.json",
		"registry/modules/../outside.json",
		`registry\modules\outside.json`,
		"dist/modules/index.json",
	}

	for _, recordPath := range paths {
		t.Run(recordPath, func(t *testing.T) {
			t.Parallel()

			_, _, err := materializeFiles(t.TempDir(), []publish.File{{
				Path:    recordPath,
				Content: []byte("unsafe\n"),
			}})
			if err == nil {
				t.Fatal("expected path to be rejected")
			}
		})
	}
}

func TestMaterializeFilesRejectsDuplicatePaths(t *testing.T) {
	t.Parallel()

	record := publish.File{
		Path:    "registry/modules/montferret/xml/manifest.json",
		Content: []byte("manifest\n"),
	}
	_, _, err := materializeFiles(t.TempDir(), []publish.File{record, record})
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("error = %v, want duplicate-path error", err)
	}
}

func TestParseOptions(t *testing.T) {
	t.Parallel()

	actual, err := parseOptions([]string{
		"--source-root", "/source",
		"--barn-root", "/barn",
		"--tag", "modules/xml/v1.2.3",
	}, io.Discard)
	if err != nil {
		t.Fatalf("parse options: %v", err)
	}
	expected := options{sourceRoot: "/source", barnRoot: "/barn", tag: "modules/xml/v1.2.3"}
	if actual != expected {
		t.Fatalf("options = %#v, want %#v", actual, expected)
	}
}

func assertFileContent(t *testing.T, root, name, expected string) {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	if string(content) != expected {
		t.Fatalf("%s content = %q, want %q", name, content, expected)
	}
}
