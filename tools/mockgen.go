package tools

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
	"github.com/rkennedy/magehelper"
)

const mockgenImport = "github.com/golang/mock/mockgen"

// MockgenReflectTask is a [mg.Fn] implementation that runs the mockgen utility to generate mock objects for given types
// in a package.
type MockgenReflectTask struct {
	mockgenBin  string
	packageName string
	types       []string
	args        []string
	deps        []any
}

var _ mg.Fn = &MockgenReflectTask{}

// MockLibrary returns a [mg.Fn] that represents the task of creating mocks for the named library using mockgen's
// reflect mode. It will create a new package by the same name but prefixed with "mock_" and then invoke mockgen with
// the given arguments, storing the result in the new directory. For example, to mock the [io.ReaderAt] and [io.Writer]
// interfaces, specify a dependency like this:
//
//	mg.CtxDeps(ctx,
//	    MockLibrary("bin/mockgen", "io", "ReaderAt", "Writer"),
//	)
//
// To mock the [github.com/logrusorgru/aurora/v3.Aurora] interface, specify a dependency like this:
//
//	mg.CtxDeps(ctx,
//	    MockLibrary("bin/mockgen", "github.com/logrusorgru/aurora/v3", "Aurora"),
//	)
//
// To specify that mocking a library has additional dependencies besides mockgen itself, call [Deps] on the returned
// MockgenReflectTask.
func MockLibrary(mockgenBin, pkg string, types ...string) *MockgenReflectTask {
	return &MockgenReflectTask{
		mockgenBin:  mockgenBin,
		packageName: pkg,
		types:       types,
	}
}

// Args sets the additional command-line arguments that are passed to mockgen. The mockLibrary will always pass
// -destination and -package, and it will include the package import path and list of interfaces automatically.
func (fn *MockgenReflectTask) Args(args ...string) *MockgenReflectTask {
	fn.args = args
	return fn
}

// Deps sets any additional magefile dependencies that mocking the current library might require.
func (fn *MockgenReflectTask) Deps(deps ...any) *MockgenReflectTask {
	fn.deps = deps
	return fn
}

// Name implements [mg.Fn].
func (fn MockgenReflectTask) Name() string {
	return fmt.Sprintf("Mockgen %s", fn.packageName)
}

// ID implements [mg.Fn].
func (fn MockgenReflectTask) ID() string {
	return fmt.Sprintf("magehelper %s %v", fn.mockgenBin, fn.types)
}

func (fn MockgenReflectTask) getInputs() (mockPackageName string, files []string, err error) {
	pkg, ok := magehelper.Packages[fn.packageName]
	if ok {
		// It's a local package.

		// We can add dependencies on the source files of that package, although we don't know precisely which
		// source files truly define the interfaces we're mocking.
		for _, file := range pkg.GoFiles {
			files = append(files, path.Join(pkg.Dir, file))
		}
		return "mock_" + pkg.Name, files, nil
	}
	// It's not a local package.
	packageName, err := sh.Output(mg.GoCmd(), "list", "-f", "{{.Name}}", fn.packageName)
	if err != nil {
		return "", nil, err
	}
	return "mock_" + packageName, nil, nil
}

func (fn MockgenReflectTask) getArgList(dest, mockPackageName string) []string {
	args := []string{
		"-destination", dest,
		// mockgen incorrectly guesses the package name based on the name of the directory it lives in. For example, the
		// default package name for github.com/logrusorgru/aurora/v3 ends up as "mock_v3" even though the package is
		// really named aurora. Therefore, we explicitly tell mockgen what we want the name of the mock package to be.
		"-package", mockPackageName,
	}
	args = append(args, fn.args...)
	return append(args, fn.packageName, strings.Join(fn.types, ","))
}

// Run implements [mg.Run].
func (fn MockgenReflectTask) Run(ctx context.Context) error {
	mg.CtxDeps(ctx, append(
		fn.deps,
		Install(fn.mockgenBin, mockgenImport),
		magehelper.LoadDependencies,
	)...)

	mockPackageName, files, err := fn.getInputs()
	if err != nil {
		return err
	}
	dest := path.Join(mockPackageName, "mocks.go")
	files = append(files, fn.mockgenBin)

	needsUpdate, err := target.Dir(dest, files...)
	if err != nil || !needsUpdate {
		return err
	}

	return sh.RunV(fn.mockgenBin, fn.getArgList(dest, mockPackageName)...)
}
