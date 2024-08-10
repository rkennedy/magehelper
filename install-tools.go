package magehelper

import (
	"context"
	"debug/buildinfo"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

func configuredModuleVersion(thisDir, module string) (string, error) {
	c := exec.Command(mg.GoCmd(),
		"list",
		"-f", "{{.Module.Version}}",
		module,
	)
	c.Dir = thisDir
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	output, err := c.Output()
	if err != nil {
		return "", err
	}
	listOutput := strings.TrimSuffix(string(output), "\n")
	logV("module %s version %s\n", module, listOutput)
	return listOutput, nil
}

func installModule(thisDir, module, bin string) error {
	gobin, err := filepath.Abs(filepath.Dir(bin))
	if err != nil {
		return err
	}
	logV("Installing %s to %s\n", module, gobin)
	c := exec.Command(mg.GoCmd(), "install", module)
	c.Env = append(os.Environ(), "GOBIN="+gobin)
	c.Dir = thisDir
	c.Stdout, c.Stderr = os.Stdout, os.Stderr
	c.Stdin = os.Stdin
	return c.Run()
}

// InstallTask is an interface that extends [mg.Fn] for tasks that install tools based on module versions recorded in
// go.mod. The ModDir function tells the task where to find the go.mod file governing the desired tool version. It
// should receive the name of the directory where go.mod is found.
type InstallTask interface {
	mg.Fn
	ModDir(dir string) InstallTask
}

type regularInstallTask struct {
	modDir string
	bin    string
	module string
}

var _ mg.Fn = &regularInstallTask{}

func (tool *regularInstallTask) ID() string {
	return fmt.Sprintf("magehelper install %s", tool.bin)
}

func (tool *regularInstallTask) Name() string {
	return fmt.Sprintf("Install %s", tool.bin)
}

func (tool *regularInstallTask) Run(context.Context) error {
	fileVersion, err := currentFileVersion(tool.bin)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	moduleVersion, err := configuredModuleVersion(tool.modDir, tool.module)
	if err != nil {
		return err
	}

	if fileVersion == moduleVersion {
		logV("Command %s is up to date.\n", tool.bin)
		return nil
	}
	return installModule(tool.modDir, tool.module, tool.bin)
}

// ModDir instructs the task where to find the go.mod file that governs the Magefile. If the magefiles are in the same
// directory as the main project, then the default blank value will be fine, but if the magefile is in a subdirectory,
// such as magefiles, and that subdirectory has its own go.mod, then magehelper requires this value so that it can know
// which version of a tool to install.
func (tool *regularInstallTask) ModDir(dir string) InstallTask {
	tool.modDir = dir
	return tool
}

const golangciLintImport = "github.com/golangci/golangci-lint/cmd/golangci-lint"

// Install returns a [mg.Fn] object suitable for using with [mg.Deps] and similar. When resolved, the object will
// install the given module to the given binary location, using the version of the module declared in go.mod. If the
// target file already exists, but has a different version, it will be replaced. If the go.mod file that controls the
// tool version is not the same as the go.mod for the project being built, then call [MagefileModuleDir] to specify
// what directory to find the right go.mod file in.
func Install(bin, module string) InstallTask {
	if module == golangciLintImport {
		task := errorInstallTask(module)
		return &task
	}
	return &regularInstallTask{bin: bin, module: module}
}

// errorInstallTask unconditionally reports an error because the given tool isn't supposed to be installed via "go
// install."
type errorInstallTask string

func (module *errorInstallTask) ID() string {
	return fmt.Sprintf("magehelper install %s", *module)
}

func (module *errorInstallTask) Name() string {
	return fmt.Sprintf("Install %s", *module)
}

func (module *errorInstallTask) Run(context.Context) error {
	return fmt.Errorf("cannot install module %s: tool isn't supposed to be installed via go-install", string(*module))
}

func (module *errorInstallTask) ModDir(string) InstallTask {
	return module
}
