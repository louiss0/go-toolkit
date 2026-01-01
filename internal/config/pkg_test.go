package config_test

import (
	"os"
	"path/filepath"

	"github.com/louiss0/cobra-cli-template/internal/config"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("IsKnownSite", func() {
	assert := assert.New(GinkgoT())

	DescribeTable("matches known sites",
		func(site string, expected bool) {
			assert.Equal(expected, config.IsKnownSite(site))
		},
		Entry("accepts github.com", "github.com", true),
		Entry("accepts gitlab.com", "gitlab.com", true),
		Entry("accepts bitbucket.org", "bitbucket.org", true),
		Entry("rejects unknown hosts", "example.com", false),
	)
})

var _ = Describe("IsValidSite", func() {
	assert := assert.New(GinkgoT())

	DescribeTable("requires a dot in the hostname",
		func(site string, expected bool) {
			assert.Equal(expected, config.IsValidSite(site))
		},
		Entry("accepts github.com", "github.com", true),
		Entry("rejects githubcom", "githubcom", false),
	)
})

var _ = Describe("ResolveUser", func() {
	assert := assert.New(GinkgoT())

	It("reads user.name from a provider config file", func() {
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "gitconfig")

		content := []byte("[user]\n\tname = Lou Name\n")
		err := os.WriteFile(configPath, content, 0o644)
		assert.NoError(err)

		values := config.Values{
			Providers: []config.ProviderConfig{
				{Name: "gitlab", Path: configPath},
			},
		}

		user, err := config.ResolveUser("", values, "gitlab.com")

		assert.NoError(err)
		assert.Equal("Lou Name", user)
	})
})
