# Mage helper

[![Go Reference](https://pkg.go.dev/badge/github.com/rkennedy/magehelper.svg)](https://pkg.go.dev/github.com/rkennedy/magehelper)

The _magehelper_ package provides package information for use in a Magefile build.

* It defines a rule for installing tool programs during the build.
* It reads dependency information from the current project to help determine prerequisites between build targets.
* It defines a rule for running the Revive linter on the code.

# Usage

```bash
go get github.com/rkennedy/magehelper
```

See this project's own magefile for example usage.
