package tools

import (
	"context"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/rkennedy/magehelper"
)

// Revive runs the given revive binary and uses the given configuration file to lint all the files in the current
// project. If revive is not installed, it will be installed using the version configured in go.mod.
func Revive(ctx context.Context, reviveBin string, config string) error {
	mg.SerialCtxDeps(ctx,
		ToolDep(reviveBin, "github.com/mgechev/revive"),
		magehelper.LoadDependencies)
	pkg, err := magehelper.BasePackage()
	if err != nil {
		return err
	}
	args := append([]string{
		"-formatter", "unix",
		"-config", config,
		"-set_exit_status",
		"./...",
	}, magehelper.Packages[pkg].IndirectGoFiles()...)
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
