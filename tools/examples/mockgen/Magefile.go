//go:build mage

// This magefile demonstrates using magehelper's Stringer tool to generate code
// during a build. Refer to the Generate target.
package main

import (
	"context"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/rkennedy/magehelper"
	"github.com/rkennedy/magehelper/tools"
)

var (
	program = filepath.Join("bin", "example")
)

// Generate updates generated code.
func Generate(ctx context.Context) {
	mg.CtxDeps(ctx,
		tools.Mockgen("subdir"),
		tools.Mockgen("."),
	)
}

// Build builds the example program.
func Build(ctx context.Context) error {
	mg.CtxDeps(ctx, Generate)
	return magehelper.Build(ctx, program)
}

// Test runs the example, confirming that Mockgen has run and produced the
// expected constants..
func Test(ctx context.Context) error {
	mg.CtxDeps(ctx, Build)
	return sh.RunV(program)
}

// All runs the test targets.
func All(ctx context.Context) {
	mg.SerialCtxDeps(ctx, Test)
}
