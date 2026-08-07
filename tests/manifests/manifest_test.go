package manifests_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/goccy/go-yaml"
)

const (
	expectedAuthor        = "MontFerret Authors"
	expectedCompatibility = ">=2.0.0-alpha.43 <3.0.0"
	expectedLicense       = "Apache-2.0"
	expectedModuleCount   = 17
	expectedRepositoryURL = "https://github.com/MontFerret/contrib"
	manifestFilename      = "ferret.yaml"
)

type (
	manifestMetadata struct {
		Name          string                 `yaml:"name"`
		Namespace     string                 `yaml:"namespace"`
		License       string                 `yaml:"license"`
		Authors       []authorMetadata       `yaml:"authors"`
		Documentation string                 `yaml:"documentation"`
		Repository    *repositoryMetadata    `yaml:"repository"`
		Compatibility *compatibilityMetadata `yaml:"compatibility"`
		Dependencies  []any                  `yaml:"dependencies"`
		Links         map[string]string      `yaml:"links"`
		Keywords      []string               `yaml:"keywords"`
		Categories    []string               `yaml:"categories"`
		Exports       *exportsMetadata       `yaml:"exports"`
	}

	authorMetadata struct {
		Name string `yaml:"name"`
	}

	repositoryMetadata struct {
		URL       string `yaml:"url"`
		Directory string `yaml:"directory"`
	}

	compatibilityMetadata struct {
		Ferret string `yaml:"ferret"`
	}

	exportsMetadata struct {
		Namespaces []namespaceMetadata `yaml:"namespaces"`
		Dialects   []string            `yaml:"dialects"`
	}

	namespaceMetadata struct {
		Name      string   `yaml:"name"`
		Functions []string `yaml:"functions"`
		Types     []string `yaml:"types"`
		Constants []string `yaml:"constants"`
	}
)

var legacyManifestFilenames = map[string]struct{}{
	"ferret-module.yaml": {},
	"ferret.module.yaml": {},
	"module.yaml":        {},
}

func TestModuleManifests(t *testing.T) {
	modulesRoot := filepath.Join("..", "..", "modules")
	moduleDirs := make([]string, 0, expectedModuleCount)
	manifestDirs := make(map[string]struct{}, expectedModuleCount)
	legacyManifests := make([]string, 0)

	err := filepath.WalkDir(modulesRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			return nil
		}

		switch entry.Name() {
		case "go.mod":
			moduleDirs = append(moduleDirs, filepath.Dir(path))
		case manifestFilename:
			manifestDirs[filepath.Dir(path)] = struct{}{}
		default:
			if _, legacy := legacyManifestFilenames[entry.Name()]; legacy {
				legacyManifests = append(legacyManifests, path)
			}
		}

		return nil
	})

	if err != nil {
		t.Fatalf("discover module manifests: %v", err)
	}

	if len(moduleDirs) != expectedModuleCount {
		t.Errorf("discovered %d modules, want %d", len(moduleDirs), expectedModuleCount)
	}

	if len(manifestDirs) != expectedModuleCount {
		t.Errorf("discovered %d manifests, want %d", len(manifestDirs), expectedModuleCount)
	}
	if len(legacyManifests) != 0 {
		sort.Strings(legacyManifests)
		t.Errorf("obsolete module manifests remain: %v", legacyManifests)
	}

	for dir := range manifestDirs {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil {
			t.Errorf("manifest %s is not beside a go.mod", filepath.Join(dir, manifestFilename))
		}
	}

	sort.Strings(moduleDirs)
	seenNames := make(map[string]string, expectedModuleCount)
	seenNamespaces := make(map[string]string, expectedModuleCount)

	for _, dir := range moduleDirs {
		rel, err := filepath.Rel(modulesRoot, dir)
		if err != nil {
			t.Fatalf("resolve module path for %s: %v", dir, err)
		}

		modulePath := filepath.ToSlash(rel)

		t.Run(modulePath, func(t *testing.T) {
			if _, exists := manifestDirs[dir]; !exists {
				t.Fatalf("missing %s", filepath.Join(dir, manifestFilename))
			}

			if _, err := os.Stat(filepath.Join(dir, "README.md")); err != nil {
				t.Errorf("documentation target is missing: %v", err)
			}

			data, err := os.ReadFile(filepath.Join(dir, manifestFilename))
			if err != nil {
				t.Fatalf("read manifest metadata: %v", err)
			}

			var manifest manifestMetadata
			err = yaml.Unmarshal(data, &manifest)
			if err != nil {
				t.Fatalf("decode manifest metadata: %v", err)
			}

			wantName := "montferret/" + filepath.Base(dir)
			if manifest.Name != wantName {
				t.Errorf("name = %q, want %q", manifest.Name, wantName)
			}

			recordUnique(t, "name", seenNames, manifest.Name, modulePath)
			recordUnique(t, "namespace", seenNamespaces, manifest.Namespace, modulePath)

			wantDocumentation := fmt.Sprintf(
				"https://github.com/MontFerret/contrib/tree/main/modules/%s/",
				modulePath,
			)

			if manifest.Documentation != wantDocumentation {
				t.Errorf("documentation = %q, want %q", manifest.Documentation, wantDocumentation)
			}

			if manifest.Repository == nil {
				t.Fatal("repository metadata is missing")
			}
			if manifest.Repository.URL != expectedRepositoryURL {
				t.Errorf("repository URL = %q, want %q", manifest.Repository.URL, expectedRepositoryURL)
			}
			wantDirectory := "modules/" + modulePath
			if manifest.Repository.Directory != wantDirectory {
				t.Errorf("repository directory = %q, want %q", manifest.Repository.Directory, wantDirectory)
			}

			if manifest.License != expectedLicense {
				t.Errorf("license = %q, want %q", manifest.License, expectedLicense)
			}

			if len(manifest.Authors) != 1 || manifest.Authors[0].Name != expectedAuthor {
				t.Errorf("authors = %#v, want exactly %q", manifest.Authors, expectedAuthor)
			}

			if manifest.Compatibility == nil || manifest.Compatibility.Ferret != expectedCompatibility {
				t.Errorf("compatibility = %#v, want Ferret %q", manifest.Compatibility, expectedCompatibility)
			}

			if manifest.Exports == nil || len(manifest.Exports.Namespaces) != 1 {
				t.Fatalf("exports = %#v, want exactly one namespace", manifest.Exports)
			}

			exported := manifest.Exports.Namespaces[0]
			if exported.Name != manifest.Namespace {
				t.Errorf("exported namespace = %q, want %q", exported.Name, manifest.Namespace)
			}

			if len(exported.Functions) == 0 {
				t.Error("exported functions must not be empty")
			}

			if len(exported.Types) != 0 || len(exported.Constants) != 0 || len(manifest.Exports.Dialects) != 0 {
				t.Errorf("unexpected non-function exports: %#v", manifest.Exports)
			}

			if len(manifest.Dependencies) != 0 || len(manifest.Links) != 0 {
				t.Error("manifest contains dependencies or links outside the agreed v1 surface")
			}

			if len(manifest.Keywords) == 0 {
				t.Error("manifest keywords must not be empty")
			}

			if len(manifest.Categories) < 2 {
				t.Error("manifest must contain at least two categories")
			}
		})
	}
}

func recordUnique(t *testing.T, field string, seen map[string]string, value, modulePath string) {
	t.Helper()

	if previous, exists := seen[value]; exists {
		t.Errorf("%s %q is also used by module %s", field, value, previous)
		return
	}

	seen[value] = modulePath
}
