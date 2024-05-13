package magehelper_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rkennedy/magehelper"
)

func TestBasePackage(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	g.Expect(magehelper.BasePackage()).To(Equal("github.com/rkennedy/magehelper"))
}
