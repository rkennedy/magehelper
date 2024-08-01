package magehelper_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/rkennedy/magehelper"
)

var _ = Describe("BasePackage", func() {
	It("returns this package's name", func() {
		Expect(magehelper.BasePackage()).To(Equal("github.com/rkennedy/magehelper"))
	})
})
