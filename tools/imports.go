package tools

import (
	"context"
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const goimportsImport = "golang.org/x/tools/cmd/goimports"

type importTask struct {
	goimportsBin string
}

func (fn *importTask) ID() string {
	return fmt.Sprintf("magehelper run %s", fn.goimportsBin)
}

func (fn *importTask) Name() string {
	return fmt.Sprintf("Goimports (%s)", fn.goimportsBin)
}

func (fn *importTask) Run(ctx context.Context) error {
	mg.CtxDeps(ctx, Install(fn.goimportsBin, goimportsImport))
	return sh.RunV(fn.goimportsBin, "-w", "-l", ".")
}

// Goimports returns a [mg.Fn] object suitable for using with [mg.Deps] and similar. When resolved, the object will run
// the given goimports binary in the project directory. If goimports is not installed, it will be installed using the
// version configured in go.mod.
func Goimports(goimportsBin string) mg.Fn {
	return &importTask{goimportsBin}
}
