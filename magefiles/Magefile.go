// This magefile determines how to build and test the project.
package main

import (
	"context"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/rkennedy/magehelper"
	"github.com/rkennedy/magehelper/tools"
)

// thisDir is the name of the directory, relative to the main module directory, where _this_ module and its go.mod file
// live.
const thisDir = "magefiles"

func goimportsBin() string {
	return filepath.Join("bin", "goimports")
}

func reviveBin() string {
	return filepath.Join("bin", "revive")
}

// Tidy cleans the go.mod file.
func Tidy(context.Context) error {
	return sh.RunV(mg.GoCmd(), "mod", "tidy")
}

// Imports formats the code and updates the import statements.
func Imports(ctx context.Context) error {
	mg.SerialCtxDeps(ctx,
		tools.Goimports(goimportsBin()).ModDir(thisDir),
		Tidy,
	)
	return nil
}

// Lint performs static analysis on all the code in the project.
func Lint(ctx context.Context) error {
	mg.SerialCtxDeps(ctx,
		Imports,
		tools.Revive(reviveBin(), "revive.toml").ModDir(thisDir),
	)
	return nil
}

// Test runs unit tests.
func Test(ctx context.Context) error {
	return magehelper.Test().Run(ctx)
}

// All runs the test and lint targets.
func All(ctx context.Context) {
	mg.SerialCtxDeps(ctx,
		Lint,
		Test,
	)
}
