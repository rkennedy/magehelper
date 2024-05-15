package tools

import (
	"context"
	"debug/buildinfo"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

func logV(s string, args ...any) {
	if mg.Verbose() {
		_, _ = fmt.Printf(s, args...)
	}
}

func currentFileVersion(bin string) (string, error) {
	binInfo, err := buildinfo.ReadFile(bin)
	if err != nil {
		// Either file doesn't exist or we couldn't read it. Either way, we want to install it.
		logV("%v\n", err)
		if err := sh.Rm(bin); err != nil {
			return "", err
		}
		return "", err
	}
	logV("%s version %s\n", bin, binInfo.Main.Version)
	return binInfo.Main.Version, nil
}

func configuredModuleVersion(module string) (string, error) {
	listOutput, err := sh.Output(
		mg.GoCmd(),
		"list",
		"-f", "{{.Module.Version}}",
		module,
	)
	if err != nil {
		return "", err
	}
	logV("module %s version %s\n", module, listOutput)
	return listOutput, nil
}

func installModule(module, bin string) error {
	gobin, err := filepath.Abs(filepath.Dir(bin))
	if err != nil {
		return err
	}
	logV("Installing %s to %s\n", module, gobin)
	return sh.RunWithV(
		map[string]string{
			"GOBIN": gobin,
		},
		mg.GoCmd(),
		"install",
		module,
	)
}

type installTask struct {
	bin    string
	module string
}

var _ mg.Fn = &installTask{}

func (tool *installTask) ID() string {
	return fmt.Sprintf("magehelper install %s", tool.bin)
}

func (tool *installTask) Name() string {
	return fmt.Sprintf("Install %s", tool.bin)
}

func (tool *installTask) Run(context.Context) error {
	fileVersion, err := currentFileVersion(tool.bin)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	moduleVersion, err := configuredModuleVersion(tool.module)
	if err != nil {
		return err
	}

	if fileVersion == moduleVersion {
		logV("Command %s is up to date.\n", tool.bin)
		return nil
	}
	return installModule(tool.module, tool.bin)
}

const golangciLintImport = "github.com/golangci/golangci-lint/cmd/golangci-lint"

// Install returns a [mg.Fn] object suitable for using with [mg.Deps] and similar. When resolved, the object will
// install the given module to the given binary location, using the version of the module declared n go.mod. If the
// target file already exists, but has a different version, it will be replaced.
func Install(bin, module string) mg.Fn {
	if module == golangciLintImport {
		return mg.F(InstallToolError, module)
	}
	return &installTask{bin, module}
}

// InstallToolError unconditionally reports an error because the given tool isn't supposed to be installed via "go
// install."
func InstallToolError(module string) error {
	return fmt.Errorf("cannot install module %s: tool isn't supposed to be installed via go-install", module)
}
