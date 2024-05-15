package tools

import (
	"context"
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/rkennedy/magehelper"
)

const reviveImport = "github.com/mgechev/revive"

type reviveTask struct {
	reviveBin string
	config    string
}

var _ mg.Fn = &reviveTask{}

func (fn *reviveTask) ID() string {
	return fmt.Sprintf("magehelper run %s", fn.reviveBin)
}

func (*reviveTask) Name() string {
	return "Revive lint"
}

func (fn *reviveTask) Run(ctx context.Context) error {
	mg.CtxDeps(ctx,
		Install(fn.reviveBin, reviveImport),
		magehelper.LoadDependencies,
	)
	pkg, err := magehelper.BasePackage()
	if err != nil {
		return err
	}
	args := append([]string{
		"-formatter", "unix",
		"-config", fn.config,
		"-set_exit_status",
		"./...",
	}, magehelper.Packages[pkg].IndirectGoFiles()...)
	return sh.RunV(
		fn.reviveBin,
		args...,
	)
}

// Revive returns a [mg.Fn] object suitable for using with [mg.Deps] and similar. When resolved, the object will run the
// given revive binary and uses the given configuration file to lint all the files in the current project. If revive is
// not installed, it will be installed using the version configured in go.mod.
func Revive(bin, config string) mg.Fn {
	return &reviveTask{bin, config}
}
