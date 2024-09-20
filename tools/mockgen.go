package tools

import (
	"context"
	"fmt"
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
	"github.com/rkennedy/magehelper/iters"
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

type mockDefinition struct {
	SourcePackage string
	External      bool
	Types         []string
}

func (def *mockDefinition) OutputPackageName(basePackage string) string {
	if def.External {
		return basePackage + "_test"
	}
	return basePackage
}

func loadRecords(dir string) ([]mockDefinition, error) {
	in, err := os.Open(filepath.Join(dir, "mockgen.yaml"))
	if err != nil {
		return nil, err
	}
	defer in.Close()

	decoder := yaml.NewDecoder(in)
	recs := map[string]mockgenRec{}
	err = decoder.Decode(&recs)
	if err != nil {
		return nil, err
	}
	return slices.Collect(iters.MapTransform(maps.All(recs), func(pkgName string, rec mockgenRec) mockDefinition {
		return mockDefinition{
			SourcePackage: pkgName,
			External:      rec.External,
			Types:         rec.Types,
		}
	})), err
}

func (fn *MockgenTask) fanOutPackages(ctx context.Context, defs []mockDefinition) {
	var wg sync.WaitGroup
	defer wg.Wait()

	for _, def := range defs {
		wg.Add(1)
		go fn.mockPackage(ctx, &wg, def)
	}
}

// Run implements [mg.Fn].
func (fn *MockgenTask) Run(ctx context.Context) error {
	dir, err := filepath.Abs(fn.dir)
	if err != nil {
		return err
	}
	fn.dir = dir

	recs, err := loadRecords(fn.dir)
	if err != nil {
		return err
	}

	fn.fanOutPackages(ctx, recs)
	return nil
}

func outputFileIfNeedsBuild(
	ctx context.Context,
	directory, sourcePackage string,
) (needsBuild bool, outputFile string, err error) {
	outFileName, files, err := outputAndInputs(ctx, directory, sourcePackage)
	if err != nil {
		return false, "", err
	}
	needsBuild, err = target.Path(outFileName, files...)
	if err != nil {
		return false, "", err
	}
	if !needsBuild {
		magehelper.LogV("File %s is up to date.\n", outFileName)
		return false, "", nil
	}
	return true, outFileName, nil
}

func (fn *MockgenTask) mockPackage(ctx context.Context, wg *sync.WaitGroup, def mockDefinition) error {
	defer wg.Done()

	needsBuild, outFileName, err := outputFileIfNeedsBuild(ctx, fn.dir, def.SourcePackage)
	if err != nil {
		return err
	}
	if !needsBuild {
		return nil
	}

	return fn.mockSinglePackage(ctx, outFileName, def)
}

// outputAndInputs determines the full name and path of the file to be generated, as well as the files that contribute
// to its generation. For a non-local package, that's just mockgen.yaml, but for a package that's part of the same
// project, the inputs include the source files for that project as well.
func outputAndInputs(ctx context.Context, dir, packageName string) (targetGoName string, files []string, err error) {
	mg.CtxDeps(ctx, magehelper.LoadDependencies)

	files = []string{filepath.Join(dir, "mockgen.yaml")}
	pkg, ok := magehelper.Packages[packageName]
	if ok {
		// It's a local package.

		// We can add dependencies on the source files of that package, although we don't know precisely which
		// source files truly define the interfaces we're mocking.
		files = slices.AppendSeq(files, iters.SliceTransform(slices.Values(pkg.GoFiles), func(file string) string {
			return filepath.Join(pkg.Dir, file)
		}))

		targetGoName = fmt.Sprintf("mock_%s_test.go", pkg.Name)
	} else {
		// It's not a local package.
		var pkgName string
		pkgName, err = sh.Output(mg.GoCmd(), "list", "-f", "{{.Name}}", packageName)
		targetGoName = fmt.Sprintf("mock_%s_test.go", pkgName)
	}
	return filepath.Join(dir, targetGoName), files, err
}

func (fn *MockgenTask) mockSinglePackage(
	ctx context.Context,
	outFileName string,
	def mockDefinition,
) error {
	pkgForDir, ok := iters.SliceSelectFirst(maps.Values(magehelper.Packages), func(pkg magehelper.Package) bool {
		return pkg.Dir == fn.dir
	})
	if !ok {
		return fmt.Errorf("No package found for directory %s", fn.dir)
	}

	mg.CtxDeps(ctx,
		magehelper.Install(fn.mockgenBin, mockgenImport).ModDir(fn.modDir),
	)
	return sh.RunV(
		fn.mockgenBin,
		"-destination", outFileName,
		"-package", def.OutputPackageName(pkgForDir.Name),
		// TODO? "-self_package", "???",
		def.SourcePackage,
		strings.Join(def.Types, ","),
	)
}
