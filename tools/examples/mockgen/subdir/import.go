// Package subdir shows how a mockgen definition in a subdirectory provides different mock objects than those in the
// base directory. The mock definition calls to mock Aurora, so it's appropriate that this be where that type's package
// be imported. It's not required that it be imported here, though, just so long as it's mentioned in go.mod.
package subdir

//revive:disable:blank-imports _Using_ the package is not important for this example.
import _ "github.com/logrusorgru/aurora/v3"
