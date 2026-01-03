package config_test

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/louiss0/cobra-cli-template/internal/config"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("ResolveConfigPath", func() {
	assert := assert.New(GinkgoT())

	It("uses the provided config path when set", func() {
		configPath := filepath.Join(GinkgoT().TempDir(), "config.toml")

		resolved := config.ResolveConfigPath(configPath)

		assert.Equal(configPath, resolved)
	})

	It("uses a repo local gtk-config.toml when present", func() {
		tempDir := GinkgoT().TempDir()
		localPath := filepath.Join(tempDir, "gtk-config.toml")

		err := os.WriteFile(localPath, []byte("site = \"github.com\"\n"), 0o644)
		assert.NoError(err)

		currentDir, err := os.Getwd()
		assert.NoError(err)

		err = os.Chdir(tempDir)
		assert.NoError(err)
		defer func() {
			_ = os.Chdir(currentDir)
		}()

		resolved := config.ResolveConfigPath("")

		assert.Equal(localPath, resolved)
	})

	It("falls back to the default config location", func() {
		tempDir := GinkgoT().TempDir()
		GinkgoT().Setenv("XDG_CONFIG_HOME", tempDir)

		currentDir, err := os.Getwd()
		assert.NoError(err)

		err = os.Chdir(tempDir)
		assert.NoError(err)
		defer func() {
			_ = os.Chdir(currentDir)
		}()

		resolved := config.ResolveConfigPath("")

		assert.Equal(filepath.Join(tempDir, "go-toolkit", "gtk-config.toml"), resolved)
	})
})

var _ = Describe("ConfigLoadSave", func() {
	assert := assert.New(GinkgoT())

	It("fails when loading with an empty path", func() {
		_, err := config.Load("")

		assert.Error(err)
	})

	It("returns empty values when the config file is missing", func() {
		path := filepath.Join(GinkgoT().TempDir(), "missing.toml")

		values, err := config.Load(path)

		assert.NoError(err)
		assert.Equal(config.Values{}, values)
	})

	It("round-trips config values", func() {
		path := filepath.Join(GinkgoT().TempDir(), "config.toml")
		values := config.Values{
			User: "lou",
			Site: "github.com",
			Providers: []config.ProviderConfig{
				{Name: "gitlab", Path: "/tmp/gitlab"},
			},
		}

		err := config.Save(path, values)
		assert.NoError(err)

		loaded, err := config.Load(path)
		assert.NoError(err)
		assert.Equal(values, loaded)
	})

	It("rejects invalid user values on save", func() {
		path := filepath.Join(GinkgoT().TempDir(), "config.toml")
		values := config.Values{
			User: "bad user",
			Site: "github.com",
		}

		err := config.Save(path, values)

		assert.Error(err)
	})
})

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
		Entry("rejects leading dots", ".github.com", false),
		Entry("rejects trailing dots", "github.com.", false),
		Entry("rejects spaces", "github com", false),
		Entry("rejects githubcom", "githubcom", false),
	)
})

var _ = Describe("ResolveSite", func() {
	assert := assert.New(GinkgoT())

	DescribeTable("resolves site",
		func(flagSite, configured, expected string) {
			values := config.Values{Site: configured}

			site := config.ResolveSite(flagSite, values)

			assert.Equal(expected, site)
		},
		Entry("uses flag override", "gitlab.com", "github.com", "gitlab.com"),
		Entry("uses configured site", "", "gitlab.com", "gitlab.com"),
		Entry("defaults to github.com", "", "", "github.com"),
	)
})

var _ = Describe("KnownSites", func() {
	assert := assert.New(GinkgoT())

	It("lists the supported providers", func() {
		known := config.KnownSites()

		assert.Contains(known, "github.com")
		assert.Contains(known, "gitlab.com")
		assert.Contains(known, "bitbucket.org")
	})
})

var _ = Describe("ResolveUser", func() {
	assert := assert.New(GinkgoT())

	It("uses the flag user when provided", func() {
		values := config.Values{}

		user, err := config.ResolveUser("lou", values, "github.com")

		assert.NoError(err)
		assert.Equal("lou", user)
	})

	It("uses the configured user when available", func() {
		values := config.Values{User: "lou"}

		user, err := config.ResolveUser("", values, "github.com")

		assert.NoError(err)
		assert.Equal("lou", user)
	})

	It("returns missing user when no providers are configured", func() {
		values := config.Values{}

		_, err := config.ResolveUser("", values, "gitlab.com")

		assert.ErrorIs(err, config.ErrMissingUser)
	})

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

	It("returns an error when the provider config path is unreadable", func() {
		values := config.Values{
			Providers: []config.ProviderConfig{
				{Name: "gitlab", Path: filepath.Join(GinkgoT().TempDir(), "missing")},
			},
		}

		_, err := config.ResolveUser("", values, "gitlab.com")

		assert.Error(err)
		assert.False(errors.Is(err, config.ErrMissingUser))
	})
})
