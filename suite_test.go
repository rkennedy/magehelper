package magehelper_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMagehelperSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Magehelper")
}
