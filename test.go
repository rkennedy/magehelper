package magehelper

import (
	"context"
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

func getDependencies(
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

func buildTestCommandLine(exe string, pkg string, tags ...string) []string {
	args := []string{
		"test",
		"-c",
		"-o", exe,
	}
	if len(tags) > 0 {
		args = append(args, "-tags", strings.Join(tags, ","))
	}
	return append(args, pkg)
}

// BuildTest builds the specified package's test.
func BuildTest(ctx context.Context, pkg string, tags ...string) error {
	mg.CtxDeps(ctx, LoadDependencies)
	deps := getDependencies(pkg, (Package).TestFiles, (Package).TestImportPackages)
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

// BuildTests build all the tests.
func BuildTests(ctx context.Context) error {
	mg.CtxDeps(ctx, LoadDependencies)
	tests := []any{}
	for _, mod := range Packages {
		tests = append(tests, mg.F(BuildTest, mod.ImportPath))
	}
	mg.CtxDeps(ctx, tests...)
	return nil
}

func runTestCommandLine(pkg string, tags []string) []string {
	args := []string{
		"test",
		"-timeout", "10s",
	}
	if len(tags) > 0 {
		args = append(args, "-tags", strings.Join(tags, ","))
	}
	return append(args, pkg)
}

// RunTest runs the specified package's tests.
func RunTest(ctx context.Context, pkg string, tags []string) error {
	mg.CtxDeps(ctx, mg.F(BuildTest, pkg, tags))

	return sh.RunV(mg.GoCmd(), runTestCommandLine(pkg, tags)...)
}

// Test runs unit tests.
func Test(ctx context.Context, tags []string) error {
	mg.CtxDeps(ctx, LoadDependencies)
	tests := []any{}
	for _, info := range Packages {
		tests = append(tests, mg.F(RunTest, info.ImportPath, tags))
	}
	mg.CtxDeps(ctx, tests...)
	return nil
}
