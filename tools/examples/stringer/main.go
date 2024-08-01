// Package stringer is an example to demonstrate bits of the magehelper package that it doesn't exercise for itself.
package main

import (
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/gomega"

	"github.com/rkennedy/magehelper/examples/stringer/subdir"
)

// SameDirectory is an enum in the same directory as the main package.
type SameDirectory int

// These are members of the enum that will be stringized.
const (
	First SameDirectory = iota
	Second
)

type test struct{}

func (*test) Helper() {}

var exitCode int

func (*test) Fatalf(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	if !strings.HasSuffix(s, "\n") {
		s = s + "\n"
	}
	_, _ = fmt.Fprint(os.Stderr, s)
	exitCode = 1
}

func main() {
	g := NewWithT(&test{})
	g.Expect(fmt.Sprintf("%s", subdir.First)).To(Equal("First"))
	g.Expect(fmt.Sprintf("%s", Second)).To(Equal("Second"))
	os.Exit(exitCode)
}
