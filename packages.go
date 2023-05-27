package magehelper

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Package represents the output from go list -json. It's based on the internal package defined in [cmd/go] and on
// [go/build.Package].
type Package struct {
	Dir        string
	ImportPath string
	Name       string
	Target     string

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

// SourceFiles returns the files that contribute to an ordinary build; these are the dependencies to check to determine
// whether the package needs to be rebuilt.
func (pkg Package) SourceFiles() []string {
	return append(pkg.GoFiles, pkg.EmbedFiles...)
}

// SourceImportPackages returns the names of other packages imported by the package.
func (pkg Package) SourceImportPackages() []string {
	return pkg.Imports
}

// TestFiles returns the files that contribute to the tests; these are the dependencies to check to determine whether
// the tests need to be rebuilt.
func (pkg Package) TestFiles() []string {
	return append(pkg.TestGoFiles, pkg.XTestGoFiles...)
}

// TestImportPackages returns the names of other packages imported by this package's tests.
func (pkg Package) TestImportPackages() []string {
	return append(pkg.TestImports, pkg.XTestImports...)
}

// IndirectGoFiles returns the files that aren't automatically selected as being part of the package proper. Contract
// with [github.com/mgechev/dots.Resolve], which will return the files that are direct members of the package, but it
// will not include other files from the same directory that belong to different packages, such as main or the _test
// package.
func (pkg Package) IndirectGoFiles() []string {
	return append(pkg.XTestGoFiles, pkg.IgnoredGoFiles...)
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
