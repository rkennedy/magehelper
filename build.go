package magehelper

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
)

func expandFiles(
	pkg Package,
	files func(pkg Package) []string,
) []string {
	var result []string
	for _, gofile := range files(pkg) {
		result = append(result, filepath.Join(pkg.Dir, gofile))
	}
	return result
}

func formatTags(tags []string) []string {
	if len(tags) > 0 {
		return []string{"-tags", strings.Join(tags, ",")}
	}
	return []string{}
}

// GetDependencies returns a list of files that the given base module depends on. The files callback should return a
// list of source files for a given package, and the imports callback should return a list of modules that the package
// imports. Use functions like [Package.SourceFiles] and [Package.SourceImportPackages].
func GetDependencies(
	baseMod string,
	files func(pkg Package) []string,
	imports func(pkg Package) []string,
) (result []string) {
	processedPackages := mapset.NewThreadUnsafeSetWithSize[string](len(Packages))
	worklist := mapset.NewSet(baseMod)

	for current, ok := worklist.Pop(); ok; current, ok = worklist.Pop() {
		if processedPackages.Add(current) {
			if pkg, ok := Packages[current]; ok {
				result = append(result, expandFiles(pkg, files)...)
				worklist.Append(imports(pkg)...)
			}
		}
	}
	return result
}

func buildBuildCommandLine(exe string, pkg string, tags []string) []string {
	args := []string{
		"build",
		"-o", exe,
	}
	args = append(args, formatTags(tags)...)
	return append(args, pkg)
}

// Build builds the current package with the given tags and writes the result to the given binary location.
func Build(ctx context.Context, exe string, tags []string) error {
	mg.CtxDeps(ctx, LoadDependencies)
	pkg, err := BasePackage()
	if err != nil {
		return err
	}
	deps := GetDependencies(pkg, Package.SourceFiles, Package.SourceImportPackages)
	newer, err := target.Path(exe, deps...)
	if err != nil || !newer {
		return err
	}
	return sh.RunV(mg.GoCmd(), buildBuildCommandLine(exe, pkg, tags)...)
}

func buildTestCommandLine(exe string, pkg string, tags ...string) []string {
	args := []string{
		"test",
		"-c",
		"-o", exe,
	}
	args = append(args, formatTags(tags)...)
	return append(args, pkg)
}

type testBuilder struct {
	pkg  string
	tags []string
}

func (tb *testBuilder) Name() string {
	return tb.ID()
}

func (tb *testBuilder) ID() string {
	return fmt.Sprintf("build-test-%s", tb.pkg)
}

func (tb *testBuilder) Run(ctx context.Context) error {
	return BuildTest(ctx, tb.pkg, tb.tags...)
}

// BuildTestDep returns a [mg.Fn] that will build the tests for the given package, subject to any given build tags.
func BuildTestDep(pkg string, tags ...string) mg.Fn {
	return &testBuilder{pkg, tags}
}

// BuildTest builds the specified package's test.
func BuildTest(ctx context.Context, pkg string, tags ...string) error {
	mg.CtxDeps(ctx, LoadDependencies)
	deps := GetDependencies(pkg, Package.TestFiles, Package.TestImportPackages)
	if len(deps) == 0 {
		return nil
	}

	info := Packages[pkg]
	exe := filepath.Join(info.Dir, info.Name+".test")

	newer, err := target.Path(exe, deps...)
	if err != nil || !newer {
		return err
	}
	return sh.RunV(
		mg.GoCmd(),
		buildTestCommandLine(exe, pkg, tags...)...,
	)
}

type allTestBuilder struct {
	tags []string
}

func (atb *allTestBuilder) Name() string {
	return atb.ID()
}

func (*allTestBuilder) ID() string {
	return "build-all-tests"
}

func (atb *allTestBuilder) Run(ctx context.Context) error {
	return BuildTests(ctx, atb.tags...)
}

// BuildTestsDep returns a [mg.Fn] that will build all the tests using the given build tags.
func BuildTestsDep(tags ...string) mg.Fn {
	return &allTestBuilder{tags}
}

// BuildTests build all the tests.
func BuildTests(ctx context.Context, tags ...string) error {
	mg.CtxDeps(ctx, LoadDependencies)
	tests := []any{}
	for _, mod := range Packages {
		tests = append(tests, BuildTestDep(mod.ImportPath, tags...))
	}
	mg.CtxDeps(ctx, tests...)
	return nil
}

func runTestCommandLine(pkg string, tags []string) []string {
	args := []string{
		"test",
		"-timeout", "10s",
	}
	args = append(args, formatTags(tags)...)
	return append(args, pkg)
}

// RunTest runs the specified package's tests.
func RunTest(ctx context.Context, pkg string, tags ...string) error {
	mg.CtxDeps(ctx, BuildTestDep(pkg, tags...))

	return sh.RunV(mg.GoCmd(), runTestCommandLine(pkg, tags)...)
}

type testRunner struct {
	pkg  string
	tags []string
}

func (tr *testRunner) Name() string {
	return tr.ID()
}

func (tr *testRunner) ID() string {
	return fmt.Sprintf("run-test-%s", tr.pkg)
}

func (tr *testRunner) Run(ctx context.Context) error {
	return RunTest(ctx, tr.pkg, tr.tags...)
}

// RunTestDep returns a [mg.Fn] that will run the tests for the given package, subject to the given build tags.
func RunTestDep(pkg string, tags ...string) mg.Fn {
	return &testRunner{pkg, tags}
}

// Test runs unit tests.
func Test(ctx context.Context, tags ...string) error {
	// It's technically not necessary to build the tests before running them; "go test" will build them anyway.
	// However, we specify BuildTests as a dependency so that _all_ the tests get built before _any_ of them start
	// running. That makes the output cleaner because lengthy test output doesn't push any build failures off the
	// top of the screen.
	mg.CtxDeps(ctx, LoadDependencies, BuildTestsDep(tags...))
	tests := []any{}
	for _, info := range Packages {
		tests = append(tests, RunTestDep(info.ImportPath, tags...))
	}
	mg.CtxDeps(ctx, tests...)
	return nil
}
