package tools_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/magefile/mage/mage"
)

var _ = Describe("Stringer", func() {
	It("builds the example project", func() {
		inv := mage.Invocation{
			Dir:    filepath.Join("examples", "stringer"),
			Stdout: GinkgoWriter,
			Stderr: GinkgoWriter,
			Args:   []string{"all"},
		}
		Expect(mage.Invoke(inv)).To(Equal(0), "Build should exit successfully.")
	})
})
