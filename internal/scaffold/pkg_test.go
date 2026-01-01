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

	It("creates a folder with an index file by default", func() {
		root := GinkgoT().TempDir()
		target := filepath.Join(root, "demo")

		err := scaffold.Create(target, scaffold.Options{
			PackageName: "demo",
			WriteIndex:  true,
		})

		assert.NoError(err)

		_, err = os.Stat(target)
		assert.NoError(err)

		content, err := os.ReadFile(filepath.Join(target, "index.go"))
		assert.NoError(err)
		assert.Contains(string(content), "package demo")
	})

	It("creates only the folder when index generation is disabled", func() {
		root := GinkgoT().TempDir()
		target := filepath.Join(root, "demo")

		err := scaffold.Create(target, scaffold.Options{
			PackageName: "demo",
			WriteIndex:  false,
			WriteReadme: false,
		})

		assert.NoError(err)

		_, err = os.Stat(filepath.Join(target, "index.go"))
		assert.Error(err)
	})

	It("creates a README when requested", func() {
		root := GinkgoT().TempDir()
		target := filepath.Join(root, "demo")

		err := scaffold.Create(target, scaffold.Options{
			PackageName: "demo",
			WriteIndex:  false,
			WriteReadme: true,
		})

		assert.NoError(err)

		content, err := os.ReadFile(filepath.Join(target, "README.md"))
		assert.NoError(err)
		assert.Contains(string(content), "# demo")
	})
})
