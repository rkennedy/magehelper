// Package magehelper_test provides tests for the magehelper module.
package magehelper_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/rkennedy/magehelper"
)

var _ = Describe("GetDependencies", func() {
	It("fetches the expected list of source files", func(ctx context.Context) {
		// It's not a great test because the magehelper package isn't very complicated, so there isn't much to exercise
		// all the corner cases.
		magehelper.LoadDependencies(ctx)

		deps := magehelper.GetDependencies("github.com/rkennedy/magehelper",
			magehelper.Package.SourceFiles,
			magehelper.Package.SourceImportPackages)
		Expect(deps).To(ContainElements(
			"build.go",
			"packages.go",
			"doc.go",
		))
	})
})
