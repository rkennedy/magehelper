//go:build mage

// This magefile determines how to build and test the project.
package main

import (
	"context"
	"fmt"
	"path"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/rkennedy/magehelper"
)

func goimportsBin() string {
	return path.Join("bin", "goimports")
}

func reviveBin() string {
	return path.Join("bin", "revive")
}

func logV(s string, args ...any) {
	if mg.Verbose() {
		_, _ = fmt.Printf(s, args...)
	}
}

// Tidy cleans the go.mod file.
func Tidy(context.Context) error {
	return sh.RunV(mg.GoCmd(), "mod", "tidy", "-go", "1.20")
}

// Imports formats the code and updates the import statements.
func Imports(ctx context.Context) error {
	mg.SerialCtxDeps(ctx,
		magehelper.ToolDep(goimportsBin(), "golang.org/x/tools/cmd/goimports"),
		Tidy,
	)
	return sh.RunV(goimportsBin(), "-w", "-l", ".")
}

// Lint performs static analysis on all the code in the project.
func Lint(ctx context.Context) error {
	mg.CtxDeps(ctx,
		Imports,
	)
	return magehelper.Revive(ctx, reviveBin(), "revive.toml")
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
