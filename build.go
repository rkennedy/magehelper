package magehelper

import (
	"context"
	"fmt"
	"iter"
	"maps"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
)

const (
	goTagOpt     = "-tags"
	ginkgoTagOpt = "--tags"
)

func formatTags(option string, tags []string) []string {
	if len(tags) > 0 {
		return []string{option, strings.Join(tags, ",")}
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
			// It's a package we haven't already processed.
			if pkg, ok := Packages[current]; ok {
				result = append(result, files(pkg)...)
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
	args = append(args, formatTags(goTagOpt, tags)...)
	return append(args, pkg)
}

// Build builds the current package with the given tags and writes the result to the given binary location.
func Build(ctx context.Context, exe string, tags ...string) error {
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
	if mg.Verbose() {
		args = append(args, "-v")
	}
	args = append(args, formatTags(goTagOpt, tags)...)
	return append(args, pkg)
}

// TestBuilder implements [mg.Fn] to build (but not run) the test binary for a single package.
type TestBuilder struct {
	pkg  string
	tags []string
}

// Name implements [mg.Fn].
func (tb *TestBuilder) Name() string {
	return tb.ID()
}

// ID implements [mg.Fn].
func (tb *TestBuilder) ID() string {
	return fmt.Sprintf("build-test-%s", tb.pkg)
}

// Run implements [mg.Fn]. If the test binary for the package needs building, then it gets built using the configured
// build tags, outputting <package-name>.test in the package director.
func (tb *TestBuilder) Run(ctx context.Context) error {
	mg.CtxDeps(ctx, LoadDependencies)
	deps := GetDependencies(tb.pkg, Package.TestFiles, Package.TestImportPackages)
	if len(deps) == 0 {
		return nil
	}

	info := Packages[tb.pkg]
	exe := info.TestBinary()

	newer, err := target.Path(exe, deps...)
	if err != nil || !newer {
		return err
	}
	return sh.RunV(mg.GoCmd(), buildTestCommandLine(exe, tb.pkg, tb.tags...)...)
}

// buildTest returns a [mg.Fn] that will build the tests for the given package, subject to any given build tags.
func buildTest(pkg string, tags ...string) *TestBuilder {
	return &TestBuilder{pkg, tags}
}

// AllGinkgoTestBuilder implements [mg.Fn] to use Ginkgo to build all the tests using build tags specified by
// [BuildTests].
type AllGinkgoTestBuilder struct {
	AllTestBuilder
	bin string
}

var _ mg.Fn = &AllGinkgoTestBuilder{}

func packageDirsNeedingBuilding() (result []string, err error) {
	for _, info := range Packages {
		deps := GetDependencies(info.ImportPath, Package.TestFiles, Package.TestImportPackages)
		if len(deps) == 0 {
			// This package has no tests.
			continue
		}

		needsBuild, err := target.Path(info.TestBinary(), deps...)
		if err != nil {
			return nil, err
		}
		if needsBuild {
			result = append(result, info.RelPath())
		}
	}
	return result, nil
}

// Run implements [mb.Fn]. It determines the list of tests in the project and runs them all on a single Ginkgo command.
func (agtb *AllGinkgoTestBuilder) Run(ctx context.Context) error {
	mg.CtxDeps(ctx,
		LoadDependencies,
		Install(agtb.bin, "github.com/onsi/ginkgo/v2/ginkgo"),
	)
	pkgs, err := packageDirsNeedingBuilding()
	if err != nil {
		return err
	}
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"build"}, formatTags(ginkgoTagOpt, agtb.tags)...)
	return sh.RunV(agtb.bin, append(args, pkgs...)...)
}

// AllTestBuilder implements [mg.Fn] to build all the tests using specified build tags.
type AllTestBuilder struct {
	tags []string
}

// Name implements [mg.Fn].
func (atb *AllTestBuilder) Name() string {
	return atb.ID()
}

// ID implements [mg.Fn].
func (*AllTestBuilder) ID() string {
	return "build-all-tests"
}

func filter[T any](src iter.Seq[T], pred func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range src {
			if pred(v) && !yield(v) {
				return
			}
		}
	}
}

// Run implements [mg.Fn]. It determines the list of tests in the project and runs them all in parallel.
func (atb *AllTestBuilder) Run(ctx context.Context) error {
	mg.CtxDeps(ctx, LoadDependencies)
	tests := []any{}
	for mod := range filter(maps.Values(Packages), Package.HasTest) {
		tests = append(tests, buildTest(mod.ImportPath, atb.tags...))
	}
	mg.CtxDeps(ctx, tests...)
	return nil
}

// UseGinkgo configures the dependency to use Ginkgo to build tests instead of plain old "go test -c." Provide the path
// and name of the ginkgo binary to run.
func (atb *AllTestBuilder) UseGinkgo(ginkgoBin string) *AllGinkgoTestBuilder {
	return &AllGinkgoTestBuilder{
		AllTestBuilder: *atb,
		bin:            ginkgoBin,
	}
}

// BuildTests returns a [mg.Fn] that will build all the tests using the given build tags.
func BuildTests(tags ...string) *AllTestBuilder {
	return &AllTestBuilder{tags}
}

func runTestCommandLine(pkg string, tags []string) []string {
	args := []string{
		"test",
		"-timeout", "10s",
	}
	if mg.Verbose() {
		args = append(args, "-v")
	}
	args = append(args, formatTags(goTagOpt, tags)...)
	return append(args, pkg)
}

// testRunner implements [mg.Fn] to build (as by [buildTest]) and run the test binary for a package using "go test."
type testRunner struct {
	pkg  string
	tags []string
}

var _ mg.Fn = &testRunner{}

// Name implements [mg.Fn].
func (tr *testRunner) Name() string {
	return tr.ID()
}

// ID implements [mg.Fn].
func (tr *testRunner) ID() string {
	return fmt.Sprintf("run-test-%s", tr.pkg)
}

// Run implements [mg.Fn]. It runs the package's test with "go test."
func (tr *testRunner) Run(ctx context.Context) error {
	mg.CtxDeps(ctx, buildTest(tr.pkg, tr.tags...))

	return sh.RunV(mg.GoCmd(), runTestCommandLine(tr.pkg, tr.tags)...)
}

// runTest returns a [mg.Fn] that will run the tests for the given package, subject to the given build tags.
func runTest(pkg string, tags ...string) *testRunner {
	return &testRunner{pkg, tags}
}

// AllGinkgoTestRunner is a [mg.Fn] that identifies all tests in the project and uses Ginkgo to build and run them.
type AllGinkgoTestRunner struct {
	AllTestRunner
	bin string
}

var _ mg.Fn = &AllGinkgoTestRunner{}

// Run implements [mg.Fn]. It uses Ginkgo to build test binaries for all applicable packages in the project (as by
// [AllGinkgoTestBuilder], and then uses Ginkgo to run them all.
func (agtr *AllGinkgoTestRunner) Run(ctx context.Context) error {
	// It's technically not necessary to build the tests before running them; "ginkgo run" would build them anyway.
	// However, we specify BuildTests as a dependency so that _all_ the tests get built before _any_ of them start
	// running. That makes the output cleaner because lengthy test output doesn't push any build failures off the
	// top of the screen.
	mg.CtxDeps(ctx,
		LoadDependencies,
		Install(agtr.bin, "github.com/onsi/ginkgo/v2/ginkgo"),
		BuildTests(agtr.tags...).UseGinkgo(agtr.bin),
	)
	args := []string{
		"run",
		"-p",
		"--timeout", "10s",
	}
	args = append(args, formatTags(ginkgoTagOpt, agtr.tags)...)
	for info := range filter(maps.Values(Packages), Package.HasTest) {
		args = append(args, info.TestBinary())
	}
	return sh.Run(agtr.bin, args...)
}

// AllTestRunner implements [mg.Fn] to identify, build, and run tests for all packages in the current project.
type AllTestRunner struct {
	tags []string
}

var _ mg.Fn = &AllTestRunner{}

// Name implements [mg.Fn].
func (atr *AllTestRunner) Name() string {
	return atr.ID()
}

// ID implements [mg.Fn].
func (*AllTestRunner) ID() string {
	return "run-all-tests"
}

// Run implements [mg.Fn] to identify, build, and run the tests for all packages in the current project. Packages
// without tests are omitted. Any tests that don't exist or that need updating will be built as with [BuildTests]. All
// tests are built before any begin running; this makes the output cleaner because any lengthy test output doesn't push
// any build failures off the top of the screen.
func (atr *AllTestRunner) Run(ctx context.Context) error {
	// It's technically not necessary to build the tests before running them; "go test" will build them anyway.
	// However, we specify BuildTests as a dependency so that _all_ the tests get built before _any_ of them start
	// running.
	mg.CtxDeps(ctx, LoadDependencies, BuildTests(atr.tags...))
	tests := []any{}
	for info := range filter(maps.Values(Packages), Package.HasTest) {
		tests = append(tests, runTest(info.ImportPath, atr.tags...))
	}
	mg.CtxDeps(ctx, tests...)
	return nil
}

// UseGinkgo configures the dependency to use Ginkgo to run the project's tests instead of running them directly as
// standalone programs.
func (atr *AllTestRunner) UseGinkgo(ginkgoBin string) *AllGinkgoTestRunner {
	return &AllGinkgoTestRunner{
		AllTestRunner: *atr,
		bin:           ginkgoBin,
	}
}

// Test returns a [mg.Fn] that identifies, builds, and runs all the tests in the project.
func Test(tags ...string) *AllTestRunner {
	return &AllTestRunner{tags: tags}
}

// LogV prints the message with [fmt.Printf] if [mg.Verbose] is true.
func LogV(msg string, args ...any) {
	if mg.Verbose() {
		_, _ = fmt.Printf(msg, args...)
	}
}
