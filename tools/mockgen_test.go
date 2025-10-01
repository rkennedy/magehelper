package tools_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"iter"
	"os"
	"path/filepath"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/magefile/mage/mage"
)

var thisDir = filepath.Join("examples", "mockgen")

func typeDecls(node ast.Node) iter.Seq[*ast.TypeSpec] {
	return func(yield func(*ast.TypeSpec) bool) {
		// Inspect doesn't offer a way to abort the inspection, so we'll just keep track of whether yield has requested
		// to stop, and then we can at least stop descending any further.
		canceled := false
		ast.Inspect(node, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.File:
				return !canceled
			case *ast.GenDecl:
				return !canceled
			case *ast.TypeSpec:
				canceled = !yield(n)
				return false
			}
			return false
		})
	}
}

var _ = Describe("Mockgen", Ordered, func() {
	BeforeAll(func() {
		By("building the example project")
		inv := mage.Invocation{
			Dir:    thisDir,
			Stdout: GinkgoWriter,
			Stderr: GinkgoWriter,
			Args:   []string{"generate"},
		}
		Expect(mage.Invoke(inv)).To(Equal(0), "Build should exit successfully.")
	})

	It("produces a valid root source file", func() {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(thisDir, "mock_io_test.go"), nil,
			parser.SkipObjectResolution)
		Expect(tree, err).NotTo(BeNil())

		Expect(tree).To(HaveField("Name.Name", Equal("main_test")))

		Expect(ast.FilterFile(tree, func(s string) bool {
			return s == "MockReaderAt"
		})).To(BeTrue())

		types := slices.Collect(typeDecls(tree))
		var structType ast.StructType
		Expect(types).To(SatisfyAll(
			HaveLen(1),
			ContainElement(HaveField("Type", BeAssignableToTypeOf(&structType))),
		))
	})

	It("produces a valid subdirectory file", func() {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(thisDir, "subdir", "mock_aurora_test.go"), nil,
			parser.SkipObjectResolution)
		Expect(tree, err).NotTo(BeNil())

		Expect(tree).To(HaveField("Name.Name", Equal("subdir")))

		Expect(ast.FilterFile(tree, func(s string) bool {
			return s == "MockAurora"
		})).To(BeTrue())

		types := slices.Collect(typeDecls(tree))
		var structType ast.StructType
		Expect(types).To(SatisfyAll(
			HaveLen(1),
			ContainElement(HaveField("Type", BeAssignableToTypeOf(&structType))),
		))
	})

	AfterAll(func() {
		// delete generated files
		os.Remove(filepath.Join(thisDir, "mock_io_test.go"))
		os.Remove(filepath.Join(thisDir, "subdir", "mock_aurora_test.go"))
	})
})
