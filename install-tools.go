package magehelper

import (
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
	logV("module version %s\n", listOutput)
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

// ToolDep returns a [mg.Fn] object suitable for using with [mg.Deps] and similar. When resolved, the object will
// install the given module to the given binary location, just like [InstallTool].
func ToolDep(bin, module string) mg.Fn {
	return mg.F(InstallTool, bin, module)
}

// InstallTool installs the given module at the given location if the file at that location either doesn't exist or
// doesn't have the same version as the version of the module configured in go.mod.
func InstallTool(bin, module string) error {
	fileVersion, err := currentFileVersion(bin)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	moduleVersion, err := configuredModuleVersion(module)
	if err != nil {
		return err
	}

	if fileVersion == moduleVersion {
		logV("Command is up to date.\n")
		return nil
	}
	return installModule(module, bin)
}
