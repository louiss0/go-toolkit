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

	It("accepts the common VCS hosts", func() {
		assert.True(config.IsKnownSite("github.com"))
		assert.True(config.IsKnownSite("gitlab.com"))
		assert.True(config.IsKnownSite("bitbucket.org"))
	})

	It("rejects unknown hosts", func() {
		assert.False(config.IsKnownSite("example.com"))
	})
})

var _ = Describe("IsValidSite", func() {
	assert := assert.New(GinkgoT())

	It("requires a dot in the hostname", func() {
		assert.True(config.IsValidSite("github.com"))
		assert.False(config.IsValidSite("githubcom"))
	})
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
