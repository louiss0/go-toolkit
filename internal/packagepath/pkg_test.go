package packagepath_test

import (
	"errors"

	"github.com/louiss0/cobra-cli-template/custom_errors"
	"github.com/louiss0/cobra-cli-template/internal/packagepath"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Packagepath", func() {
	assert := assert.New(GinkgoT())

	Describe("NormalizePackageName", func() {
		It("replaces spaces with underscores", func() {
			name := packagepath.NormalizePackageName("my package")

			assert.Equal("my_package", name)
		})

		It("collapses repeated spaces", func() {
			name := packagepath.NormalizePackageName("  multiple   spaces  ")

			assert.Equal("multiple_spaces", name)
		})
	})

	Describe("ResolveModulePath", func() {
		It("keeps fully qualified module paths", func() {
			path, err := packagepath.ResolveModulePath("github.com/acme/tool", "github.com", "lou")

			assert.NoError(err)
			assert.Equal("github.com/acme/tool", path)
		})

		It("accepts numeric module segments when the path is complete", func() {
			path, err := packagepath.ResolveModulePath("4/r/7", "github.com", "lou")

			assert.NoError(err)
			assert.Equal("4/r/7", path)
		})

		It("accepts full module paths with a version suffix", func() {
			path, err := packagepath.ResolveModulePath("github.com/onsi/ginkgo/v2", "github.com", "lou")

			assert.NoError(err)
			assert.Equal("github.com/onsi/ginkgo/v2", path)
		})

		It("adds the site when given user and package", func() {
			path, err := packagepath.ResolveModulePath("acme/tool", "github.com", "lou")

			assert.NoError(err)
			assert.Equal("github.com/acme/tool", path)
		})

		It("uses the registered user when only a package is provided", func() {
			path, err := packagepath.ResolveModulePath("tool", "github.com", "lou")

			assert.NoError(err)
			assert.Equal("github.com/lou/tool", path)
		})

		It("errors when the user is missing for a short package", func() {
			_, err := packagepath.ResolveModulePath("tool", "github.com", "")

			assert.True(errors.Is(err, packagepath.ErrMissingUser))
		})

		It("errors when the site is invalid for a short package", func() {
			_, err := packagepath.ResolveModulePath("tool", "githubcom", "lou")

			assert.True(errors.Is(err, custom_errors.InvalidInput))
		})

		It("errors on invalid path structures", func() {
			_, err := packagepath.ResolveModulePath("a/b/c/d", "github.com", "lou")

			assert.True(errors.Is(err, custom_errors.InvalidInput))
		})
	})
})
