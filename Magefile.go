//go:build mage

// This magefile determines how to build and test the project.
package main

import (
	"context"
	"path"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/rkennedy/magehelper"
	"github.com/rkennedy/magehelper/tools"
)

func goimportsBin() string {
	return path.Join("bin", "goimports")
}

func reviveBin() string {
	return path.Join("bin", "revive")
}

// Tidy cleans the go.mod file.
func Tidy(context.Context) error {
	return sh.RunV(mg.GoCmd(), "mod", "tidy", "-go", "1.20")
}

// Imports formats the code and updates the import statements.
func Imports(ctx context.Context) error {
	mg.SerialCtxDeps(ctx,
		tools.Goimports(goimportsBin()),
		Tidy,
	)
	return nil
}

// Lint performs static analysis on all the code in the project.
func Lint(ctx context.Context) error {
	mg.SerialCtxDeps(ctx,
		Imports,
		tools.Revive(reviveBin(), "revive.toml"),
	)
	return nil
}

// Test runs unit tests.
func Test(ctx context.Context) error {
	return magehelper.Test(ctx)
}

// All runs the test and lint targets.
func All(ctx context.Context) {
	mg.SerialCtxDeps(ctx,
		Lint,
		Test,
	)
}
