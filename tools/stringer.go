package tools

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
	"github.com/rkennedy/magehelper"
)

const stringerImport = "golang.org/x/tools/cmd/stringer"

// StringerTask is a [mg.Fn] implementation that runs the stringer utility to generate code for an enum type.
type StringerTask struct {
	stringerBin     string
	typeName        string
	destinationFile string
	inputFiles      []string

	modDir string
}

var _ mg.Fn = &StringerTask{}

// ID implements [mg.Fn].
func (fn *StringerTask) ID() string {
	return fmt.Sprintf("magehelper run %s(%s)", fn.stringerBin, fn.destinationFile)
}

// Name implements [mg.Fn].
func (fn *StringerTask) Name() string {
	return fmt.Sprintf("Stringer %s", fn.typeName)
}

// ModDir indicates the directory where a go.mod file exists specifying which version of the string module to use for
// this task.
func (fn *StringerTask) ModDir(dir string) *StringerTask {
	fn.modDir = dir
	return fn
}

type noInputError string

func (e noInputError) Error() string {
	return fmt.Sprintf("no input files for %s", string(e))
}

// Run implements [mg.Fn].
func (fn *StringerTask) Run(ctx context.Context) error {
	if len(fn.inputFiles) <= 0 {
		return noInputError(fn.typeName)
	}

	mg.CtxDeps(ctx, magehelper.Install(fn.stringerBin, stringerImport).ModDir(fn.modDir))

	needsUpdate, err := target.Dir(fn.destinationFile, append(fn.inputFiles, fn.stringerBin)...)
	if err != nil || !needsUpdate {
		return err
	}
	// We'll assume that all input files are in the same directory, so it's safe to select the first one.
	packageDir := filepath.Dir(fn.inputFiles[0])
	if !filepath.IsAbs(packageDir) {
		// The stringer command will interpret a bare directory as if
		// it's a package name, and then fail to load it. We need to
		// make sure it looks like a directory. We can't prepend a "."
		// like this because Join calls Clean, which removes it.
		// packageDir = filepath.Join(".", packageDir)
		// Instead, we can make the path absolute:
		packageDir, _ = filepath.Abs(packageDir)
	}

	return sh.RunV(fn.stringerBin, "-output", fn.destinationFile, "-type", fn.typeName, packageDir)
}

// Stringer returns a [mg.Fn] object suitable for using with [mg.Deps] and similar. When resolved, the object will run
// the stringer utility to generate code for the given types. and store the result in the given destination file. At
// least one input file is required. Input files are used to calculate whether the destination file is out of date
// first. The stringer utility is installed if it's not present or if it's out of date.
func Stringer(stringerBin string, typeName, destination string, inputFiles ...string) *StringerTask {
	return &StringerTask{
		stringerBin:     stringerBin,
		typeName:        typeName,
		destinationFile: destination,
		inputFiles:      inputFiles,
	}
}
