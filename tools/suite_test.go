package tools_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestToolSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Magehelper tools")
}
