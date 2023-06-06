// Package magehelper_test provides tests for the magehelper module.
package magehelper_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rkennedy/magehelper"
)

// TestGetDependencies checks that [magehelper.GetDependencies] fetches the expected list of source files. It's not a
// great test because the magehelper package isn't very complicated, so there isn't much to exercise all the corner
// cases.
func TestGetDependencies(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	magehelper.LoadDependencies(context.Background())

	deps := magehelper.GetDependencies("github.com/rkennedy/magehelper",
		(magehelper.Package).SourceFiles,
		(magehelper.Package).SourceImportPackages)
	g.Expect(deps).To(ContainElements(
		HaveSuffix("/build.go"),
		HaveSuffix("/revive.go"),
		HaveSuffix("/packages.go"),
		HaveSuffix("/doc.go"),
		HaveSuffix("/install-tools.go"),
	))
}
