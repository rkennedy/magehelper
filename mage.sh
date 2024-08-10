#!/bin/bash

set -euo pipefail

(cd magefiles; go build -o ../bin/mage mage.go)
export MAGEFILE_ENABLE_COLOR=1
exec bin/mage "$@"
