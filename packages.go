package magehelper

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"golang.org/x/mod/modfile"
)

// Package represents the output from go list -json. It's based on the internal package defined in [cmd/go] and on
// [go/build.Package].
type Package struct {
	Dir        string
	ImportPath string
	Name       string
	Target     string
	Root       string

	GoFiles        []string
	IgnoredGoFiles []string
	TestGoFiles    []string
	XTestGoFiles   []string

	EmbedFiles      []string
	TestEmbedFiles  []string
	XTestEmbedFiles []string

	Imports      []string
	TestImports  []string
	XTestImports []string
}

// flattenRel receives a sequence of slices of source file names and returns a flattened slice of all their names
// prefixed with their relative directory.
func flattenRel(base, dir string, paths ...[]string) []string {
	relPath, err := filepath.Rel(base, dir)
	if err != nil {
		relPath = dir
	}

	result := []string{}
	for _, pathGroup := range paths {
		for _, path := range pathGroup {
			result = append(result, filepath.Join(relPath, path))
		}
	}
	return result
}

// SourceFiles returns the files that contribute to an ordinary build; these are the dependencies to check to determine
// whether the package needs to be rebuilt.
func (pkg Package) SourceFiles() []string {
	return flattenRel(pkg.Root, pkg.Dir, pkg.GoFiles, pkg.EmbedFiles)
}

// SourceImportPackages returns the names of other packages imported by the package.
func (pkg Package) SourceImportPackages() []string {
	return pkg.Imports
}

// HasTest indicates whether the package has any tests. It checks whether both [TestGoFiles] and [XTestGoFiles] are
// empty.
func (pkg Package) HasTest() bool {
	return len(pkg.TestGoFiles) != 0 || len(pkg.XTestGoFiles) != 0
}

// TestFiles returns the files that contribute to the tests; these are the dependencies to check to determine whether
// the tests need to be rebuilt.
func (pkg Package) TestFiles() []string {
	return flattenRel(pkg.Root, pkg.Dir, pkg.TestGoFiles, pkg.XTestGoFiles, pkg.TestEmbedFiles, pkg.XTestEmbedFiles)
}

// TestImportPackages returns the names of other packages imported by this package's tests.
func (pkg Package) TestImportPackages() []string {
	return append(pkg.TestImports, pkg.XTestImports...)
}

// RelPath returns the path of the package relative to its root. If that path cannot be calculated, then it returns the
// package's Dir value instead.
func (pkg Package) RelPath() string {
	relPath, err := filepath.Rel(pkg.Root, pkg.Dir)
	if err != nil {
		return pkg.Dir
	}
	return relPath
}

// TestBinary returns the name and path of the package's test binary, relative to the package root. The test name is
// the package name followed by .test.
func (pkg Package) TestBinary() string {
	return filepath.Join(pkg.RelPath(), pkg.Name+".test")
}

// IndirectGoFiles returns the files that aren't automatically selected as being part of the package proper. Contrast
// with [github.com/mgechev/dots.Resolve], which will return the files that are direct members of the package, but it
// will not include other files from the same directory that belong to different packages, such as main or the _test
// package.
func (pkg Package) IndirectGoFiles() []string {
	return flattenRel(pkg.Root, pkg.Dir, pkg.XTestGoFiles, pkg.IgnoredGoFiles)
}

// Packages holds the results of the [LoadDependencies] function. This variable is only valid after that function runs.
// Use [mg.Deps] or similar to make sure dependencies are loaded before referring to this variable.
var Packages = map[string]Package{}

// LoadDependencies populates the global [Packages] variable. It's suitable for use with [mg.Deps] or [mg.CtxDeps].
func LoadDependencies(context.Context) error {
	dependencies, err := sh.Output(mg.GoCmd(), "list", "-json", "./...")
	if err != nil {
		return err
	}
	dec := json.NewDecoder(strings.NewReader(dependencies))
	for {
		var pkg Package
		switch err = dec.Decode(&pkg); err {
		case io.EOF:
			return nil
		case nil:
			Packages[pkg.ImportPath] = pkg
		default:
			return err
		}
	}
}

// BasePackage returns the full name of the package being tested. It reads go.mod, which means this is intended to be
// used on tests being run in situ from the project directory, not tests compiled and copied elsewhere to run.
func BasePackage() (string, error) {
	f, err := os.Open("go.mod")
	if err != nil {
		return "", err
	}
	defer f.Close()

	bytes, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return modfile.ModulePath(bytes), nil
}
