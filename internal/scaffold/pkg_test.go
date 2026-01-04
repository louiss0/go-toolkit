package scaffold_test

import (
	"os"
	"path/filepath"

	"github.com/louiss0/cobra-cli-template/internal/scaffold"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Scaffold", func() {
	assert := assert.New(GinkgoT())

	It("fails when package name is missing", func() {
		err := scaffold.Create(GinkgoT().TempDir(), scaffold.Options{})

		assert.Error(err)
	})

	It("fails when the folder path is missing", func() {
		err := scaffold.Create("", scaffold.Options{PackageName: "demo"})

		assert.Error(err)
	})

	It("fails when the folder path is a file", func() {
		root := GinkgoT().TempDir()
		target := filepath.Join(root, "demo")

		err := os.WriteFile(target, []byte("not a dir"), 0o644)
		assert.NoError(err)

		err = scaffold.Create(target, scaffold.Options{
			PackageName:   "demo",
			WriteRootFile: true,
		})

		assert.Error(err)
	})

	It("creates a folder with a root file by default", func() {
		root := GinkgoT().TempDir()
		target := filepath.Join(root, "demo")

		err := scaffold.Create(target, scaffold.Options{
			PackageName:   "demo",
			WriteRootFile: true,
		})

		assert.NoError(err)

		_, err = os.Stat(target)
		assert.NoError(err)

		content, err := os.ReadFile(filepath.Join(target, "demo.go"))
		assert.NoError(err)
		assert.Contains(string(content), "package demo")
	})

	It("creates only the folder when root file generation is disabled", func() {
		root := GinkgoT().TempDir()
		target := filepath.Join(root, "demo")

		err := scaffold.Create(target, scaffold.Options{
			PackageName:   "demo",
			WriteRootFile: false,
			WriteReadme:   false,
		})

		assert.NoError(err)

		_, err = os.Stat(filepath.Join(target, "demo.go"))
		assert.Error(err)
	})

	It("creates a README when requested", func() {
		root := GinkgoT().TempDir()
		target := filepath.Join(root, "demo")

		err := scaffold.Create(target, scaffold.Options{
			PackageName:   "demo",
			WriteRootFile: false,
			WriteReadme:   true,
		})

		assert.NoError(err)

		content, err := os.ReadFile(filepath.Join(target, "README.md"))
		assert.NoError(err)
		assert.Contains(string(content), "# demo")
	})

	It("creates a test file when enabled", func() {
		root := GinkgoT().TempDir()
		target := filepath.Join(root, "demo")

		err := scaffold.Create(target, scaffold.Options{
			PackageName:   "demo",
			WriteRootFile: true,
			WriteTests:    true,
		})

		assert.NoError(err)

		content, err := os.ReadFile(filepath.Join(target, "demo_test.go"))
		assert.NoError(err)
		assert.Contains(string(content), "func TestDemo(t *testing.T)")
	})
})
