package magehelper

import (
	"context"
	"io"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"golang.org/x/mod/modfile"
)

// getBasePackage returns the full name of the package being tested. It reads go.mod, which means this is intended to be
// used on tests being run in situ from the project directory, not tests compiled and copied elsewhere to run.
func getBasePackage() (string, error) {
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

// Revive runs the given revive binary and uses the given configuration file to lint all the files in the current
// project. If revive is not installed, it will be installed using the version configured in go.mod.
func Revive(ctx context.Context, reviveBin string, config string) error {
	mg.SerialCtxDeps(ctx,
		ToolDep(reviveBin, "github.com/mgechev/revive"),
		LoadDependencies)
	pkg, err := getBasePackage()
	if err != nil {
		return err
	}
	args := append([]string{
		"-formatter", "unix",
		"-config", config,
		"-set_exit_status",
		"./...",
	}, Packages[pkg].IndirectGoFiles()...)
	return sh.RunV(
		reviveBin,
		args...,
	)
}

// ReviveDep returns a [mg.Fn] object suitable for using with [mg.Deps] and similar. When resolved, the object will run
// the given revive binary and uses the given configuration file to lint all the files in the current project. If revive
// is not installed, it will be installed using the version configured in go.mod.
func ReviveDep(bin, config string) mg.Fn {
	return mg.F(Revive, bin, config)
}
