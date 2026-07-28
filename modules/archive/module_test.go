package archive

import (
	"archive/tar"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/ziflex/go-options"

	"github.com/MontFerret/contrib/modules/archive/core"
	"github.com/MontFerret/ferret/v2"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk/sdktest"
	"github.com/MontFerret/ferret/v2/pkg/source"
)

func TestNewSmoke(t *testing.T) {
	mod := New()

	if mod == nil {
		t.Fatal("expected module")
	}
	if mod.Name() != "archive" {
		t.Fatalf("expected module name archive, got %q", mod.Name())
	}
}

func TestNewRejectsNegativeLimitsAtBootstrap(t *testing.T) {
	tests := []struct {
		option Option
		field  string
		name   string
	}{
		{
			name:   "entry size",
			option: WithMaxEntrySize(-1),
			field:  "MaxEntrySize",
		},
		{
			name:   "ZIP buffer size",
			option: WithMaxZIPBufferSize(-1),
			field:  "MaxZIPBufferSize",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ferret.New(ferret.WithModules(New(test.option)))
			var validationErr options.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("expected typed module configuration error, got %v", err)
			}
			if validationErr.Field != test.field ||
				validationErr.Reason != "must be non-negative" ||
				validationErr.Value != "-1" {
				t.Fatalf("unexpected module configuration error: %#v", validationErr)
			}
		})
	}
}

func TestModuleInstallsNormalizedConfigInRunContext(t *testing.T) {
	tests := []struct {
		name     string
		options  []Option
		caller   core.Config
		expected core.Config
	}{
		{
			name: "configured limits override caller config",
			options: []Option{
				WithMaxEntrySize(7),
				WithMaxZIPBufferSize(11),
			},
			caller: core.Config{
				MaxEntrySize:     1,
				MaxZIPBufferSize: 2,
			},
			expected: core.Config{
				MaxEntrySize:     7,
				MaxZIPBufferSize: 11,
			},
		},
		{
			name: "later zero limits restore defaults",
			options: []Option{
				WithMaxEntrySize(7),
				WithMaxZIPBufferSize(11),
				WithMaxEntrySize(0),
				WithMaxZIPBufferSize(0),
			},
			caller: core.Config{
				MaxEntrySize:     1,
				MaxZIPBufferSize: 2,
			},
			expected: core.DefaultConfig(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var observed core.Config

			engine, err := ferret.New(
				ferret.WithAfterRunHook(func(ctx context.Context, _ error) error {
					var lookupErr error
					observed, lookupErr = core.ConfigFrom(ctx)

					return lookupErr
				}),
				ferret.WithModules(New(test.options...)),
			)
			if err != nil {
				t.Fatalf("create engine: %v", err)
			}
			t.Cleanup(func() {
				if err := engine.Close(); err != nil {
					t.Fatalf("close engine: %v", err)
				}
			})

			runCtx := core.WithConfig(context.Background(), test.caller)
			if _, err := engine.Run(runCtx, source.NewAnonymous("RETURN true")); err != nil {
				t.Fatalf("run engine: %v", err)
			}
			if observed != test.expected {
				t.Fatalf("unexpected run config: got %#v, want %#v", observed, test.expected)
			}
		})
	}
}

func TestModuleListReadAndExtractZIPFromFQL(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeZIPForTest(t, root, "release.zip", []archiveTestEntry{
		{name: "docs", typeflag: tar.TypeDir, mode: 0o755},
		{name: "docs/readme.txt", body: "hello archive", mode: 0o644},
		{name: "current", linkName: "docs/readme.txt", typeflag: tar.TypeSymlink, mode: 0o777},
	})

	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithFSRoot(root),
		ferret.WithRuntimeParam("source", runtime.NewString("release.zip")),
	)
	output, err := harness.Run(context.Background(), `
		LET entries = ARCHIVE::LIST(@source)
		LET text = ARCHIVE::READ(@source, "docs/readme.txt", { as: "string" })
		LET binary = ARCHIVE::READ(@source, "docs/readme.txt")
		LET missing = ARCHIVE::READ(@source, "missing.txt", { missing: "none" })
		LET extracted = ARCHIVE::EXTRACT(@source, "dist", {
			include: ["docs", "docs/*", "current"],
			links: "skip"
		})
		RETURN {
			entries: entries,
			text: text,
			binaryType: TYPENAME(binary),
			binaryText: TO_STRING(binary),
			missing: missing,
			extracted: extracted
		}
	`)
	if err != nil {
		t.Fatalf("run archive workflow: %v", err)
	}

	var actual struct {
		Text       string `json:"text"`
		BinaryType string `json:"binaryType"`
		BinaryText string `json:"binaryText"`
		Missing    any    `json:"missing"`
		Entries    []struct {
			LinkName       *string `json:"linkName"`
			CompressedSize *int64  `json:"compressedSize"`
			Name           string  `json:"name"`
			Mode           string  `json:"mode"`
			Format         string  `json:"format"`
			IsSymlink      bool    `json:"isSymlink"`
		} `json:"entries"`
		Extracted []struct {
			Path    *string `json:"path"`
			Reason  *string `json:"reason"`
			Name    string  `json:"name"`
			Skipped bool    `json:"skipped"`
		} `json:"extracted"`
	}
	if err := json.Unmarshal(output.Content, &actual); err != nil {
		t.Fatalf("decode FQL result: %v", err)
	}

	if actual.Text != "hello archive" || actual.BinaryType != "Binary" ||
		actual.BinaryText != "hello archive" || actual.Missing != nil {
		t.Fatalf("unexpected READ results: %#v", actual)
	}
	if len(actual.Entries) != 3 || actual.Entries[1].Name != "docs/readme.txt" ||
		actual.Entries[1].Mode != "0644" || actual.Entries[1].Format != "zip" ||
		actual.Entries[1].CompressedSize == nil {
		t.Fatalf("unexpected LIST results: %#v", actual.Entries)
	}
	if !actual.Entries[2].IsSymlink || actual.Entries[2].LinkName == nil ||
		*actual.Entries[2].LinkName != "docs/readme.txt" {
		t.Fatalf("unexpected ZIP link metadata: %#v", actual.Entries[2])
	}
	if actual.Entries[1].LinkName != nil {
		t.Fatalf("expected a regular entry linkName to be NONE, got %#v", actual.Entries[1].LinkName)
	}
	if len(actual.Extracted) != 3 || !actual.Extracted[2].Skipped ||
		actual.Extracted[2].Path != nil || actual.Extracted[2].Reason == nil ||
		*actual.Extracted[2].Reason != "link" {
		t.Fatalf("unexpected extraction results: %#v", actual.Extracted)
	}

	var raw struct {
		Entries   []map[string]json.RawMessage `json:"entries"`
		Extracted []map[string]json.RawMessage `json:"extracted"`
	}
	if err := json.Unmarshal(output.Content, &raw); err != nil {
		t.Fatalf("decode raw FQL result: %v", err)
	}
	assertObjectFields(t, raw.Entries[0],
		"compressedSize", "format", "isDir", "isRegular", "isSymlink",
		"linkName", "modTime", "mode", "name", "size",
	)
	assertObjectFields(t, raw.Extracted[0],
		"isDir", "name", "path", "reason", "size", "skipped",
	)

	content, err := os.ReadFile(filepath.Join(root, "dist", "docs", "readme.txt"))
	if err != nil {
		t.Fatalf("read extracted file: %v", err)
	}
	if string(content) != "hello archive" {
		t.Fatalf("unexpected extracted content %q", content)
	}
}

func TestModuleReadsTARGZByContentDetection(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTARForTest(t, root, "payload.bin", true, []archiveTestEntry{
		{name: "manifest.json", body: `{"ok":true}`, mode: 0o600},
	})

	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithFSRoot(root),
	)
	output, err := harness.Run(context.Background(), `
		LET entry = FIRST(ARCHIVE::LIST("payload.bin"))
		RETURN {
			format: entry.format,
			compressed: entry.compressedSize,
			value: ARCHIVE::READ("payload.bin", "manifest.json", { as: "string" })
		}
	`)
	if err != nil {
		t.Fatalf("run TAR.GZ workflow: %v", err)
	}

	var actual struct {
		Compressed any    `json:"compressed"`
		Format     string `json:"format"`
		Value      string `json:"value"`
	}
	if err := json.Unmarshal(output.Content, &actual); err != nil {
		t.Fatalf("decode TAR.GZ result: %v", err)
	}
	if actual.Format != "tar.gz" || actual.Value != `{"ok":true}` || actual.Compressed != nil {
		t.Fatalf("unexpected TAR.GZ result: %#v", actual)
	}
}

func TestModuleRejectsUnsafeEntryBeforeWriting(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeZIPForTest(t, root, "unsafe.zip", []archiveTestEntry{
		{name: "safe.txt", body: "safe"},
		{name: "../evil.txt", body: "evil"},
	})

	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithFSRoot(root),
	)
	_, err := harness.Run(context.Background(), `
		RETURN ARCHIVE::EXTRACT("unsafe.zip", "dist")
	`)
	if err == nil || !strings.Contains(err.Error(), "unsafe archive entry path") {
		t.Fatalf("expected unsafe path error, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "dist")); !os.IsNotExist(statErr) {
		t.Fatalf("expected no destination side effects, got %v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(root, "evil.txt")); !os.IsNotExist(statErr) {
		t.Fatalf("expected no traversal output, got %v", statErr)
	}
}

func TestModuleRejectsDestinationSymlink(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeZIPForTest(t, root, "release.zip", []archiveTestEntry{{name: "index.html", body: "body"}})
	if err := os.Mkdir(filepath.Join(root, "target"), 0o755); err != nil {
		t.Fatalf("create target: %v", err)
	}
	if err := os.Symlink("target", filepath.Join(root, "dist")); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}

	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithFSRoot(root),
	)
	_, err := harness.Run(context.Background(), `
		RETURN ARCHIVE::EXTRACT("release.zip", "dist")
	`)
	if err == nil || !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("expected destination symlink error, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "target", "index.html")); !os.IsNotExist(statErr) {
		t.Fatalf("expected symlink target to remain untouched, got %v", statErr)
	}
}

func TestModuleRejectsConfiguredEntryLimit(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeZIPForTest(t, root, "large.zip", []archiveTestEntry{{name: "large.txt", body: "12345"}})

	harness := sdktest.New(t,
		ferret.WithModules(New(WithMaxEntrySize(4))),
		ferret.WithFSRoot(root),
	)
	_, err := harness.Run(context.Background(), `
		RETURN ARCHIVE::READ("large.zip", "large.txt")
	`)
	if err == nil || !strings.Contains(err.Error(), "exceeding the limit") {
		t.Fatalf("expected limit error, got %v", err)
	}
}

func TestModuleRejectsUnknownOptions(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeZIPForTest(t, root, "release.zip", []archiveTestEntry{{name: "index.html", body: "body"}})

	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithFSRoot(root),
	)
	_, err := harness.Run(context.Background(), `
		RETURN ARCHIVE::LIST("release.zip", { unknown: true })
	`)
	if err == nil {
		t.Fatalf("expected unknown option error, got %v", err)
	}
}

func TestModuleReportsMissingFilesystem(t *testing.T) {
	t.Parallel()

	harness := sdktest.New(t, ferret.WithModules(New()))
	_, err := harness.Run(context.Background(), `
		RETURN ARCHIVE::LIST("release.zip")
	`)
	if err == nil || !strings.Contains(err.Error(), "root is not configured") {
		t.Fatalf("expected disabled filesystem error, got %v", err)
	}
}

func TestModuleHonorsReadOnlyFilesystem(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeZIPForTest(t, root, "release.zip", []archiveTestEntry{{name: "index.html", body: "body"}})

	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithFSRoot(root),
		ferret.WithFSReadOnly(),
	)
	_, err := harness.Run(context.Background(), `
		RETURN ARCHIVE::EXTRACT("release.zip", "dist")
	`)
	if err == nil || !strings.Contains(err.Error(), "read-only") {
		t.Fatalf("expected read-only error, got %v", err)
	}
}

func TestModuleExtractionFiltersAndOverwrite(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTARForTest(t, root, "site.tar", false, []archiveTestEntry{
		{name: "public/index.html", body: "new"},
		{name: "public/app.map", body: "map"},
		{name: "private/secret.txt", body: "secret"},
	})
	if err := os.MkdirAll(filepath.Join(root, "dist", "public"), 0o755); err != nil {
		t.Fatalf("create destination: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "dist", "public", "index.html"), []byte("old"), 0o600); err != nil {
		t.Fatalf("write existing destination: %v", err)
	}

	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithFSRoot(root),
	)
	_, err := harness.Run(context.Background(), `
		RETURN ARCHIVE::EXTRACT("site.tar", "dist", {
			overwrite: true,
			include: ["public/*"],
			exclude: ["public/*.map"]
		})
	`)
	if err != nil {
		t.Fatalf("extract filtered TAR: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(root, "dist", "public", "index.html"))
	if err != nil {
		t.Fatalf("read overwritten output: %v", err)
	}
	if string(content) != "new" {
		t.Fatalf("unexpected overwritten output %q", content)
	}
	if _, err := os.Stat(filepath.Join(root, "dist", "public", "app.map")); !os.IsNotExist(err) {
		t.Fatalf("expected excluded map to be absent, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "dist", "private", "secret.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected non-included file to be absent, got %v", err)
	}
}

func TestModuleExtractionRejectsExistingFileWithoutOverwrite(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeZIPForTest(t, root, "release.zip", []archiveTestEntry{{name: "index.html", body: "new"}})
	if err := os.Mkdir(filepath.Join(root, "dist"), 0o755); err != nil {
		t.Fatalf("create destination: %v", err)
	}
	target := filepath.Join(root, "dist", "index.html")
	if err := os.WriteFile(target, []byte("old"), 0o600); err != nil {
		t.Fatalf("write existing destination: %v", err)
	}

	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithFSRoot(root),
	)
	_, err := harness.Run(context.Background(), `
		RETURN ARCHIVE::EXTRACT("release.zip", "dist")
	`)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected existing destination error, got %v", err)
	}

	content, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("read preserved destination: %v", readErr)
	}
	if string(content) != "old" {
		t.Fatalf("expected destination to remain unchanged, got %q", content)
	}
}

func TestModuleExtractionWithNoEligibleEntriesDoesNotCreateDestination(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeZIPForTest(t, root, "release.zip", []archiveTestEntry{{name: "index.html", body: "body"}})

	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithFSRoot(root),
	)
	output, err := harness.Run(context.Background(), `
		RETURN ARCHIVE::EXTRACT("release.zip", "dist", { include: ["assets/*"] })
	`)
	if err != nil {
		t.Fatalf("extract with no eligible entries: %v", err)
	}
	if string(output.Content) != "[]" {
		t.Fatalf("expected an empty result, got %s", output.Content)
	}
	if _, statErr := os.Stat(filepath.Join(root, "dist")); !os.IsNotExist(statErr) {
		t.Fatalf("expected no destination side effects, got %v", statErr)
	}
}

func TestModuleLinkErrorPreflightsBeforeWriting(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTARForTest(t, root, "links.tar", false, []archiveTestEntry{
		{name: "safe.txt", body: "safe"},
		{name: "current", linkName: "safe.txt", typeflag: tar.TypeLink},
	})

	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithFSRoot(root),
	)
	_, err := harness.Run(context.Background(), `
		RETURN ARCHIVE::EXTRACT("links.tar", "dist", { links: "error" })
	`)
	if err == nil || !strings.Contains(err.Error(), "links are disabled") {
		t.Fatalf("expected link policy error, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "dist")); !os.IsNotExist(statErr) {
		t.Fatalf("expected link preflight to avoid destination writes, got %v", statErr)
	}
}

func TestModuleRejectsSpecialTAREntryBeforeWriting(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTARForTest(t, root, "special.tar", false, []archiveTestEntry{
		{name: "safe.txt", body: "safe"},
		{name: "pipe", typeflag: tar.TypeFifo},
	})

	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithFSRoot(root),
	)
	_, err := harness.Run(context.Background(), `
		RETURN ARCHIVE::EXTRACT("special.tar", "dist")
	`)
	if err == nil || !strings.Contains(err.Error(), "unsupported special file type") {
		t.Fatalf("expected special entry error, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "dist")); !os.IsNotExist(statErr) {
		t.Fatalf("expected special entry preflight to avoid writes, got %v", statErr)
	}
}

func TestModuleCreateDirsFalseRequiresDestination(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeZIPForTest(t, root, "release.zip", []archiveTestEntry{{name: "index.html", body: "body"}})

	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithFSRoot(root),
	)
	_, err := harness.Run(context.Background(), `
		RETURN ARCHIVE::EXTRACT("release.zip", "missing", { createDirs: false })
	`)
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected missing directory error, got %v", err)
	}
}

func TestModuleRejectsDuplicateExtractionPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeZIPForTest(t, root, "duplicate.zip", []archiveTestEntry{
		{name: "same.txt", linkName: "target.txt", typeflag: tar.TypeSymlink},
		{name: "same.txt", body: "first"},
		{name: "same.txt", body: "second"},
		{name: "only-link", linkName: "target.txt", typeflag: tar.TypeSymlink},
	})

	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithFSRoot(root),
	)
	_, err := harness.Run(context.Background(), `
		RETURN ARCHIVE::EXTRACT("duplicate.zip", "dist")
	`)
	if err == nil || !strings.Contains(err.Error(), "duplicate extraction path") {
		t.Fatalf("expected duplicate path error, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "dist")); !os.IsNotExist(statErr) {
		t.Fatalf("expected duplicate preflight to avoid writes, got %v", statErr)
	}
}

func TestModuleReadUsesFirstRegularDuplicate(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeZIPForTest(t, root, "duplicate.zip", []archiveTestEntry{
		{name: "same.txt", body: "first"},
		{name: "same.txt", body: "second"},
	})

	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithFSRoot(root),
	)
	output, err := harness.Run(context.Background(), `
		RETURN ARCHIVE::READ("duplicate.zip", "same.txt", { as: "string" })
	`)
	if err != nil {
		t.Fatalf("read duplicate: %v", err)
	}

	var actual string
	if err := json.Unmarshal(output.Content, &actual); err != nil {
		t.Fatalf("decode duplicate result: %v", err)
	}
	if actual != "first" {
		t.Fatalf("expected first duplicate, got %q", actual)
	}

	_, err = harness.Run(context.Background(), `
		RETURN ARCHIVE::READ("duplicate.zip", "only-link")
	`)
	if err == nil {
		t.Fatalf("expected non-regular entry error, got %v", err)
	}
}

func assertObjectFields(t *testing.T, actual map[string]json.RawMessage, expected ...string) {
	t.Helper()

	actualFields := make([]string, 0, len(actual))
	for field := range actual {
		actualFields = append(actualFields, field)
	}
	sort.Strings(actualFields)
	sort.Strings(expected)
	if strings.Join(actualFields, ",") != strings.Join(expected, ",") {
		t.Fatalf("unexpected object fields: got %v, want %v", actualFields, expected)
	}
}
