//go:build mage

// This magefile determines how to build and test the project.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/rkennedy/magehelper"
	"golang.org/x/mod/modfile"
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
	mg.SerialCtxDeps(ctx, magehelper.ToolDep(goimportsBin(), "golang.org/x/tools/cmd/goimports"), Tidy)
	return sh.RunV(goimportsBin(), "-w", "-l", ".")
}

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

// Lint performs static analysis on all the code in the project.
func Lint(ctx context.Context) error {
	mg.SerialCtxDeps(ctx,
		Imports,
		magehelper.ToolDep(reviveBin(), "github.com/mgechev/revive"),
		magehelper.LoadDependencies)
	pkg, err := getBasePackage()
	if err != nil {
		return err
	}
	args := append([]string{
		"-formatter", "unix",
		"-config", "revive.toml",
		"-set_exit_status",
		"./...",
	}, magehelper.Packages[pkg].IndirectGoFiles()...)
	return sh.RunWithV(
		map[string]string{
			"REVIVE_FORCE_COLOR": "1",
		},
		reviveBin(),
		args...,
	)
}

// All runs the build, test, and lint targets.
func All(ctx context.Context) {
	mg.SerialCtxDeps(ctx, Lint)
}
