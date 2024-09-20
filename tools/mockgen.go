package tools

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
	"github.com/rkennedy/magehelper"
	"gopkg.in/yaml.v3"
)

const mockgenImport = "github.com/golang/mock/mockgen"

// MockgenTask is a Mage task that generates mock types for code in a particular directory.
type MockgenTask struct {
	mockgenBin string
	modDir     string

	dir string
}

var _ mg.Fn = &MockgenTask{}

// Mockgen returns a [mg.Fn] that represents the task of creating mocks for use by the given directory. In that
// directory should be a file, mockgen.yaml, containing the packages and types that dir's code needs to use mock
// versions of. The YAML should have the structure of the following type:
//
//	var yamlStructure map[string]struct {
//	    external bool
//	    types    []string
//	}
//
// That is, the YAML should be an object whose keys are the names of the packages whose types need to be mocked. The
// values are objects indicating the list of types in the package that should be mocked. The external field indicates
// whether the generated code should be in the package of the current directory (false) or in the "_test" package
// (true).
//
// The task will run mockgen with the given path and binary name, and it will be installed according to the version
// requested in the active go.mod, specified by [Mockgen.ModDir].
//
// When mockgen.yaml calls for mocking types from another package in the same module, that's referred to as a "local"
// package, and all that package's source files will be included as dependencies for the generated mock source file,
// along with mockgen.yaml itself. If the mocked types come from a non-local package (i.e., a Go built-in package or a
// third-party package), then only mockgen.yaml is a dependency. When dependencies are newer than the generated mock
// source file, then the source will gen regenerated.
//
// To mock the [io.ReaderAt], [io.WriterAt], and [github.com/logrusorgru/aurora/v3.Aurora] interfaces, specify a
// mockgen.yaml file like this:
//
//	io:
//	  external: true
//	  types:
//	  - ReaderAt
//	  - Writer
//	github.com/logrusorgru/aurora/v3:
//	  external: true
//	  types:
//	  - Aurora
func Mockgen(mockgenBin, dir string) *MockgenTask {
	return &MockgenTask{
		mockgenBin: mockgenBin,
		dir:        dir,
	}
}

// ModDir sets the directory where this task will look for a go.mod file that specifies which version of mockgen this
// project uses. If not specified, it will look in the current working directory, which is usually the project root, but
// a common alternative is to have a separate go.mod in the magefile directory so that build requirements don't leak
// into the main project dependencies. ModDir returns the MockgenTask.
func (fn *MockgenTask) ModDir(dir string) *MockgenTask {
	fn.modDir = dir
	return fn
}

// Name implements [mg.Fn].
func (fn MockgenTask) Name() string {
	return fmt.Sprintf("Mockgen directory %s", fn.dir)
}

// ID implements [mg.Fn].
func (fn MockgenTask) ID() string {
	return fmt.Sprintf("magehelper mockgen %s", fn.dir)
}

// Each directory that _uses_ mocks will have a mockgen.yaml file listing all the packages and types that it needs mocks
// from.
type mockgenRec struct {
	External bool     `yaml:"external"`
	Types    []string `yaml:"types"`
}

// Run implements [mg.Fn].
func (fn *MockgenTask) Run(ctx context.Context) error {
	dir, err := filepath.Abs(fn.dir)
	if err != nil {
		return err
	}
	fn.dir = dir
	in, err := os.Open(filepath.Join(fn.dir, "mockgen.yaml"))
	if err != nil {
		return err
	}
	defer in.Close()

	decoder := yaml.NewDecoder(in)
	recs := map[string]mockgenRec{}
	err = decoder.Decode(&recs)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	for packageName, rec := range recs {
		wg.Add(1)
		go fn.mockPackage(ctx, &wg, packageName, rec)
	}
	return nil
}

func firstThat[T any](items iter.Seq[T], pred func(T) bool) (T, error) {
	for t := range items {
		if pred(t) {
			return t, nil
		}
	}
	return *new(T), errors.New("no match found")
}

func (fn *MockgenTask) mockPackage(ctx context.Context, wg *sync.WaitGroup, packageName string, rec mockgenRec) error {
	defer wg.Done()

	outFileName, needsBuild, err := shouldBuildFile(ctx, fn.dir, packageName)
	if err != nil {
		return err
	}
	if !needsBuild {
		if mg.Verbose() {
			_, _ = fmt.Printf("File %s is up to date.\n", outFileName)
		}
		return nil
	}

	pkgForDir, err := firstThat(maps.Values(magehelper.Packages), func(pkg magehelper.Package) bool {
		return pkg.Dir == fn.dir
	})
	if err != nil {
		return err
	}
	outpkg := pkgForDir.Name
	if rec.External {
		outpkg += "_test"
	}

	return (&mockgenSubtask{
		parentTask:  fn,
		outFileName: outFileName,
		outpkg:      outpkg,
		packageName: packageName,
		types:       rec.Types,
	}).Run(ctx)
}

func transform[T, U any](src iter.Seq[T], f func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for t := range src {
			if !yield(f(t)) {
				return
			}
		}
	}
}

func shouldBuildFile(ctx context.Context, dir, packageName string) (targetGoName string, needsBuild bool, err error) {
	mg.CtxDeps(ctx, magehelper.LoadDependencies)

	files := []string{filepath.Join(dir, "mockgen.yaml")}
	pkg, ok := magehelper.Packages[packageName]
	if ok {
		// It's a local package.

		// We can add dependencies on the source files of that package, although we don't know precisely which
		// source files truly define the interfaces we're mocking.
		files = slices.AppendSeq(files, transform(slices.Values(pkg.GoFiles), func(file string) string {
			return filepath.Join(pkg.Dir, file)
		}))

		targetGoName = fmt.Sprintf("mock_%s_test.go", pkg.Name)
	} else {
		// It's not a local package.
		var pkgName string
		pkgName, err = sh.Output(mg.GoCmd(), "list", "-f", "{{.Name}}", packageName)
		if err != nil {
			return "", false, err
		}
		targetGoName = fmt.Sprintf("mock_%s_test.go", pkgName)
	}
	targetGoName = filepath.Join(dir, targetGoName)

	needsBuild, err = target.Path(targetGoName, files...)
	return targetGoName, needsBuild, err
}

// mockgenSubtask is the part of a MockgenTask that runs mockgen for a single package.
type mockgenSubtask struct {
	parentTask  *MockgenTask
	outFileName string
	outpkg      string
	packageName string
	types       []string
}

var _ mg.Fn = &mockgenSubtask{}

func (fn mockgenSubtask) Name() string {
	return fmt.Sprintf("Mock %s.{%s} in %s", fn.packageName, fn.types, fn.parentTask.dir)
}

func (fn mockgenSubtask) ID() string {
	return fmt.Sprintf("magehelper mock %s.{%s}/%s", fn.packageName, fn.types, fn.parentTask.dir)
}

func (fn mockgenSubtask) Run(ctx context.Context) error {
	mg.CtxDeps(ctx,
		magehelper.Install(fn.parentTask.mockgenBin, mockgenImport).ModDir(fn.parentTask.modDir),
	)
	return sh.RunV(
		fn.parentTask.mockgenBin,
		"-destination", fn.outFileName,
		"-package", fn.outpkg,
		// TODO? "-self_package", path.Join(mustBasePackage(), fn.parentTask.dir),
		fn.packageName,
		strings.Join(fn.types, ","),
	)
}
