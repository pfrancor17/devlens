package domain_test

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// TestDomainImportsOnlyStandardLibrary enforces the hexagonal rule: the domain
// never depends on adapters, use cases or third-party packages.
func TestDomainImportsOnlyStandardLibrary(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read domain dir: %v", err)
	}

	fset := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if isThirdParty(path) {
				t.Errorf("%s imports %q; the domain may only use the standard library", name, path)
			}
		}
	}
}

// Standard library paths have no dot in their first element.
func isThirdParty(path string) bool {
	first, _, _ := strings.Cut(path, "/")
	return strings.Contains(first, ".")
}
