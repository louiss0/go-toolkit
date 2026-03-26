package validation_test

import (
	"github.com/louiss0/go-toolkit/validation"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("RequiredString", func() {
	It("returns the trimmed string", func() {
		value, err := validation.RequiredString("  lou  ", "user")

		assert.NoError(err)
		assert.Equal("lou", value)
	})

	It("rejects empty strings", func() {
		_, err := validation.RequiredString("   ", "user")

		assert.Error(err)
		assert.Contains(err.Error(), "user is required")
	})
})

var _ = Describe("ParseBool", func() {
	It("parses boolean strings", func() {
		value, err := validation.ParseBool("true", "enabled")

		assert.NoError(err)
		assert.True(value)
	})

	It("rejects invalid boolean strings", func() {
		_, err := validation.ParseBool("maybe", "enabled")

		assert.Error(err)
		assert.Contains(err.Error(), "enabled must be true or false")
	})
})

var _ = Describe("ValidateSite", func() {
	It("accepts known sites", func() {
		err := validation.ValidateSite("github.com", false, []string{"github.com", "gitlab.com"})

		assert.NoError(err)
	})

	It("rejects unknown sites when full sites are disabled", func() {
		err := validation.ValidateSite("example.com", false, []string{"github.com", "gitlab.com"})

		assert.Error(err)
		assert.Contains(err.Error(), "unsupported site")
	})
})

var _ = Describe("ParseShortPackageList", func() {
	It("accepts blank input for optional prompts", func() {
		packages, err := validation.ParseShortPackageList("   ", "packages to install")

		assert.NoError(err)
		assert.Nil(packages)
	})

	It("parses space-delimited short package paths", func() {
		packages, err := validation.ParseShortPackageList("samber/lo stretchr/testify", "packages to install")

		assert.NoError(err)
		assert.Equal([]string{"samber/lo", "stretchr/testify"}, packages)
	})

	It("parses versioned short package paths", func() {
		packages, err := validation.ParseShortPackageList("onsi/ginkgo/v2", "packages to install")

		assert.NoError(err)
		assert.Equal([]string{"onsi/ginkgo/v2"}, packages)
	})

	It("rejects commas and full module paths", func() {
		_, err := validation.ParseShortPackageList(
			"samber/lo, github.com/spf13/cobra",
			"packages to install",
		)

		assert.Error(err)
		assert.Contains(err.Error(), "packages to install must use space-separated username/package or username/package/vN entries")
	})
})

var _ = Describe("RequiredShortPackageList", func() {
	It("requires at least one short package path", func() {
		_, err := validation.RequiredShortPackageList("   ", "packages to add")

		assert.Error(err)
		assert.Contains(err.Error(), "packages to add must use space-separated username/package or username/package/vN entries")
	})
})

var _ = Describe("IsFullModulePath", func() {
	It("accepts fully-qualified module paths", func() {
		assert.True(validation.IsFullModulePath("github.com/samber/lo"))
		assert.True(validation.IsFullModulePath("rsc.io/quote"))
	})

	It("rejects short package paths", func() {
		assert.False(validation.IsFullModulePath("samber/lo"))
	})
})

var _ = Describe("IsShortPackagePath", func() {
	It("accepts owner, package, and major version segments", func() {
		assert.True(validation.IsShortPackagePath("onsi/ginkgo/v2"))
	})

	It("rejects a third segment that is not a major version", func() {
		assert.False(validation.IsShortPackagePath("onsi/ginkgo/release"))
	})
})
