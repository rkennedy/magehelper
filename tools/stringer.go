package tools

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
)

const stringerImport = "golang.org/x/tools/cmd/stringer"

// StringerTask is a [mg.Fn] implementation that runs the stringer utility to generate code for an enum type.
type StringerTask struct {
	stringerBin     string
	typeName        string
	destinationFile string
	inputFiles      []string
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

type noInputError string

func (e noInputError) Error() string {
	return fmt.Sprintf("no input files for %s", string(e))
}

// Run implements [mg.Fn].
func (fn *StringerTask) Run(ctx context.Context) error {
	if len(fn.inputFiles) <= 0 {
		return noInputError(fn.typeName)
	}

	mg.CtxDeps(ctx, Install(fn.stringerBin, stringerImport))

	dest := fn.destinationFile

	needsUpdate, err := target.Dir(dest, append(fn.inputFiles, fn.stringerBin)...)
	if err != nil || !needsUpdate {
		return err
	}
	// We'll assume that all input files are in the same directory, so it's safe to select the first one.
	packageDir := filepath.Dir(fn.inputFiles[0])

	return sh.RunV(fn.stringerBin, "-output", dest, "-type", fn.typeName, packageDir)
}

// Stringer returns a [mg.Fn] object suitable for using with [mg.Deps] and similar. When resolved, the object will run
// the stringer utility to generate code for the given types. and store the result in the given destination file. At
// least one input file is required. Input files are used to calculate whether the destination file is out of date
// first. The string utility is installed if it's not present or if it's out of date.
func Stringer(stringerBin string, typeName, destination string, inputFiles ...string) mg.Fn {
	return &StringerTask{
		stringerBin:     stringerBin,
		typeName:        typeName,
		destinationFile: destination,
		inputFiles:      inputFiles,
	}
}
