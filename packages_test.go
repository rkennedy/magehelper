package magehelper_test

import (
	"context"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/rkennedy/magehelper"
)

const thisPackage = "github.com/rkennedy/magehelper"

var _ = Describe("BasePackage", func() {
	It("returns this package's name", func() {
		Expect(magehelper.BasePackage()).To(Equal(thisPackage))
	})
})

var _ = Describe("LoadDependencies", func() {
	BeforeEach(func() {
		Expect(magehelper.LoadDependencies(context.Background())).To(Succeed(), "Should load package information")
	})

	It("includes known packages", func() {
		Expect(magehelper.Packages).To(SatisfyAll(
			HaveKey(thisPackage),
			HaveKey(path.Join(thisPackage, "tools")),
		))
	})

	Context("detects test presence", func() {
		It("in packages with tests", func() {
			pkg, ok := magehelper.Packages[thisPackage]
			Expect(ok).To(BeTrue(), "Package list should include %s", thisPackage)
			Expect(pkg.HasTest()).To(BeTrue(), "Package should have a test (i.e., this one)")
		})

		It("in packages without tests", func() {
			notest := path.Join(thisPackage, "notest")
			pkg, ok := magehelper.Packages[notest]
			Expect(ok).To(BeTrue(), "Package list should include %s", notest)
			Expect(pkg.HasTest()).To(BeFalse(), "Package should not have a test; found %#v", pkg.TestFiles())
		})
	})
})
