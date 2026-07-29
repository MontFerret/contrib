package archive

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionCodeDoesNotCallDirectHostIO(t *testing.T) {
	t.Parallel()

	banned := map[string]map[string]struct{}{
		"os": {
			"Open": {}, "OpenFile": {}, "Create": {}, "ReadFile": {}, "WriteFile": {},
			"Stat": {}, "Lstat": {}, "Mkdir": {}, "MkdirAll": {}, "Remove": {}, "RemoveAll": {},
		},
		"net":      {"Dial": {}},
		"net/http": {"Get": {}},
		"path/filepath": {
			"EvalSymlinks": {},
		},
	}

	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return parseErr
		}
		for _, spec := range file.Imports {
			importPath := strings.Trim(spec.Path.Value, `"`)
			if importPath == "net/http" || importPath == "path/filepath" || importPath == "net" {
				t.Errorf("production file %s imports prohibited package %s", path, importPath)
			}
		}

		full, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if parseErr != nil {
			return parseErr
		}
		ast.Inspect(full, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			identifier, ok := selector.X.(*ast.Ident)
			if !ok {
				return true
			}
			if names, exists := banned[identifier.Name]; exists {
				if _, prohibited := names[selector.Sel.Name]; prohibited {
					t.Errorf("production file %s calls prohibited function %s.%s", path, identifier.Name, selector.Sel.Name)
				}
			}

			return true
		})

		return nil
	})
	if err != nil {
		t.Fatalf("inspect production code: %v", err)
	}
}
