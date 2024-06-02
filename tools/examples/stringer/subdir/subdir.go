// Package subdir is a package for demonstrating how to use the Stringer tool for magehelper.
package subdir

// Subdirectory is an enum type for Stringer to process. It lives in a
// subdirectory of the main project.
type Subdirectory int

// These constants are found by stringer.
const (
	First Subdirectory = iota
	Second
)
