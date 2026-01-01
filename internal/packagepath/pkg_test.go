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
		DescribeTable("normalizes names",
			func(input, expected string) {
				name := packagepath.NormalizePackageName(input)

				assert.Equal(expected, name)
			},
			Entry("replaces spaces with underscores", "my package", "my_package"),
			Entry("collapses repeated spaces", "  multiple   spaces  ", "multiple_spaces"),
		)
	})

	Describe("ResolveModulePath", func() {
		DescribeTable("resolves module paths",
			func(input, site, user, expected string, expectedErr error) {
				path, err := packagepath.ResolveModulePath(input, site, user)

				if expectedErr != nil {
					assert.Error(err)
					assert.True(errors.Is(err, expectedErr))
					return
				}

				assert.NoError(err)
				assert.Equal(expected, path)
			},
			Entry(
				"keeps fully qualified module paths",
				"github.com/acme/tool",
				"github.com",
				"lou",
				"github.com/acme/tool",
				nil,
			),
			Entry(
				"accepts numeric module segments when the path is complete",
				"4/r/7",
				"github.com",
				"lou",
				"4/r/7",
				nil,
			),
			Entry(
				"accepts full module paths with a version suffix",
				"github.com/onsi/ginkgo/v2",
				"github.com",
				"lou",
				"github.com/onsi/ginkgo/v2",
				nil,
			),
			Entry(
				"adds the site when given user and package",
				"acme/tool",
				"github.com",
				"lou",
				"github.com/acme/tool",
				nil,
			),
			Entry(
				"uses the registered user when only a package is provided",
				"tool",
				"github.com",
				"lou",
				"github.com/lou/tool",
				nil,
			),
			Entry(
				"errors when the user is missing for a short package",
				"tool",
				"github.com",
				"",
				"",
				packagepath.ErrMissingUser,
			),
			Entry(
				"errors when the site is invalid for a short package",
				"tool",
				"githubcom",
				"lou",
				"",
				custom_errors.InvalidInput,
			),
			Entry(
				"errors on invalid path structures",
				"a/b/c/d",
				"github.com",
				"lou",
				"",
				custom_errors.InvalidInput,
			),
		)
	})
})
