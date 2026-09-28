// deps_test.go verifies AC7's boundary: tooling/management and
// tooling/distribution never import an external module, even transitively
// through their own tricell-hive/... imports, even though tooling/cli now
// depends on Bubble Tea and bubbles for its interactive interface (design.md
// "Dependencias"). It walks each package's own .go files with go/parser (imports
// only, so a syntax error elsewhere in a file can never mask a real
// import), recursing into every tricell-hive/... import's own directory
// (resolved from the module root) instead of merely skipping it — a T2 fix
// round finding: skipping those imports outright would silently miss an
// external dependency hidden two or more hops away — and fails on any
// import anywhere in that closure whose first path element contains a dot,
// a domain-shaped module path such as charm.land/... or github.com/....
package management

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestNoExternalDependencies(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{".", "../distribution"} {
		t.Run(dir, func(t *testing.T) {
			checkNoExternalImports(t, "tricell-hive", root, dir)
		})
	}
}

// testingT is the subset of *testing.T checkNoExternalImports needs.
// TestNoExternalDependencies passes a real *testing.T (which satisfies this
// structurally); TestNoExternalDependenciesCatchesTransitiveImport passes a
// recordingT instead, so it can assert on a reported violation without that
// violation also failing its own test.
type testingT interface {
	Helper()
	Fatalf(format string, args ...any)
	Errorf(format string, args ...any)
}

// checkNoExternalImports walks pkgDir's own tricell-hive/... import closure
// (every package it imports under modulePrefix, and every package those
// import, and so on, each resolved from moduleRoot), failing on any import
// anywhere in that closure whose first path element contains a dot.
// modulePrefix and moduleRoot are parameters, not hardcoded to
// "tricell-hive", so a synthetic fixture can exercise the exact same
// walker.
func checkNoExternalImports(t testingT, modulePrefix, moduleRoot, pkgDir string) {
	t.Helper()
	visited := map[string]bool{}
	var walk func(dir string)
	walk = func(dir string) {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			t.Fatalf("resolving %s: %v", dir, err)
			return
		}
		if visited[absDir] {
			return
		}
		visited[absDir] = true
		entries, err := os.ReadDir(absDir)
		if err != nil {
			t.Fatalf("reading %s: %v", absDir, err)
			return
		}
		fset := token.NewFileSet()
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
				continue
			}
			path := filepath.Join(absDir, e.Name())
			f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("parsing %s: %v", path, err)
				continue
			}
			for _, imp := range f.Imports {
				importPath, err := strconv.Unquote(imp.Path.Value)
				if err != nil {
					t.Fatalf("%s: invalid import literal %s: %v", path, imp.Path.Value, err)
					continue
				}
				if importPath == modulePrefix || strings.HasPrefix(importPath, modulePrefix+"/") {
					rel := strings.TrimPrefix(strings.TrimPrefix(importPath, modulePrefix), "/")
					walk(filepath.Join(moduleRoot, filepath.FromSlash(rel)))
					continue
				}
				first, _, _ := strings.Cut(importPath, "/")
				if strings.Contains(first, ".") {
					t.Errorf("%s imports external module %q; the whole %s/... closure reached from %s must depend only on the standard library and %s/...", path, importPath, modulePrefix, pkgDir, modulePrefix)
				}
			}
		}
	}
	walk(pkgDir)
}

// recordingT is a minimal testingT double: it inherits Helper and Fatalf
// from the embedded *testing.T (a read/parse error in the fixture below
// would still stop the test, which is correct), but records Errorf calls
// instead of letting them fail this test — TestNoExternalDependencies-
// CatchesTransitiveImport wants to assert that checkNoExternalImports
// itself reports a violation, not to actually fail from one.
type recordingT struct {
	*testing.T
	failed    bool
	errorText string
}

func (r *recordingT) Errorf(format string, args ...any) {
	r.failed = true
	r.errorText += fmt.Sprintf(format, args...) + "\n"
}

// TestNoExternalDependenciesCatchesTransitiveImport is a mutation-style
// fixture for checkNoExternalImports itself, over a synthetic two-package
// module: the entry package imports only an internal sibling, and that
// sibling is the one with the external dependency. A version of
// checkNoExternalImports that only skips modulePrefix-prefixed imports
// instead of recursing into them (the T2 fix round's finding about the
// original deps_test.go) would never look inside "sibling" and would miss
// this violation entirely; this proves the current walker does not.
func TestNoExternalDependenciesCatchesTransitiveImport(t *testing.T) {
	root := t.TempDir()
	entryDir := filepath.Join(root, "entry")
	siblingDir := filepath.Join(root, "sibling")
	if err := os.MkdirAll(entryDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(siblingDir, 0700); err != nil {
		t.Fatal(err)
	}
	entrySrc := "package entry\n\nimport _ \"synthetic/sibling\"\n"
	if err := os.WriteFile(filepath.Join(entryDir, "entry.go"), []byte(entrySrc), 0600); err != nil {
		t.Fatal(err)
	}
	siblingSrc := "package sibling\n\nimport _ \"github.com/external/pkg\"\n"
	if err := os.WriteFile(filepath.Join(siblingDir, "sibling.go"), []byte(siblingSrc), 0600); err != nil {
		t.Fatal(err)
	}

	rec := &recordingT{T: t}
	checkNoExternalImports(rec, "synthetic", root, entryDir)

	if !rec.failed {
		t.Fatal("checkNoExternalImports did not catch the transitive external import")
	}
	if !strings.Contains(rec.errorText, "github.com/external/pkg") {
		t.Fatalf("violation did not name the offending import: %s", rec.errorText)
	}
}

// TestNoExternalDependenciesAllowsCleanTransitiveImport is
// CatchesTransitiveImport's negative counterpart: an entry package that
// only reaches internal (synthetic/...) and standard-library imports,
// however deep, must report no violation.
func TestNoExternalDependenciesAllowsCleanTransitiveImport(t *testing.T) {
	root := t.TempDir()
	entryDir := filepath.Join(root, "entry")
	siblingDir := filepath.Join(root, "sibling")
	if err := os.MkdirAll(entryDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(siblingDir, 0700); err != nil {
		t.Fatal(err)
	}
	entrySrc := "package entry\n\nimport _ \"synthetic/sibling\"\n"
	if err := os.WriteFile(filepath.Join(entryDir, "entry.go"), []byte(entrySrc), 0600); err != nil {
		t.Fatal(err)
	}
	siblingSrc := "package sibling\n\nimport _ \"strings\"\n"
	if err := os.WriteFile(filepath.Join(siblingDir, "sibling.go"), []byte(siblingSrc), 0600); err != nil {
		t.Fatal(err)
	}

	rec := &recordingT{T: t}
	checkNoExternalImports(rec, "synthetic", root, entryDir)

	if rec.failed {
		t.Fatalf("checkNoExternalImports reported a false violation: %s", rec.errorText)
	}
}
