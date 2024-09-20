//go:build mage

// This magefile demonstrates using magehelper's Stringer tool to generate code
// during a build. Refer to the Generate target.
package main

import (
	"context"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/rkennedy/magehelper/tools"
)

var (
	mockgenBin = filepath.Join("bin", "mockgen")
)

// Generate updates generated code.
func Generate(ctx context.Context) {
	mg.CtxDeps(ctx,
		tools.Mockgen(mockgenBin, "subdir"),
		tools.Mockgen(mockgenBin, "."),
	)
}
