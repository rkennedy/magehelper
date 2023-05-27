# Mage helper

[![Go Reference](https://pkg.go.dev/badge/github.com/rkennedy/magehelper.svg)](https://pkg.go.dev/github.com/rkennedy/magehelper)

The _magehelper_ package provides package information for use in a Magefile build.

# Usage

```bash
go get github.com/rkennedy/magehelper
```

```go
package main

import (
	"os"

	"github.com/rkennedy/nblog"
	"golang.org/x/exp/slog"
)

func main() {
	logger := slog.New(nblog.NewHandler(os.Stdout))
	logger.Info("message")
}
```
