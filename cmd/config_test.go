package cmd_test

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/louiss0/go-toolkit/cmd"
	"github.com/louiss0/go-toolkit/internal/modindex/config"
	"github.com/louiss0/go-toolkit/internal/testhelpers"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var Config = Describe("config command", func() {
	assert := assert.New(GinkgoT())

	It("initializes config with defaults", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		_, err := testhelpers.ExecuteCmd(rootCmd, "config", "init", "--user", "lou")

		assert.NoError(err)

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Equal("lou", values.User)
		assert.Equal("github.com", values.Site)
		assert.False(values.AssureProviders)
	})

	It("prompts for config init when no flags are provided", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		promptRunner := testhelpers.NewPromptRunnerMock(
			testhelpers.PromptStep{Kind: testhelpers.PromptStepInput, Value: "lou"},
			testhelpers.PromptStep{Kind: testhelpers.PromptStepSelect, Value: "gitlab.com"},
		)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: promptRunner,
			ConfigPath:   configPath,
		})

		output, err := testhelpers.ExecuteCmd(rootCmd, "config", "init")

		assert.NoError(err)

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Equal("lou", values.User)
		assert.Equal("gitlab.com", values.Site)

		var summary map[string]any
		err = json.Unmarshal([]byte(output), &summary)
		assert.NoError(err)
		assert.Equal("gitlab.com", summary["site"])
		assert.Equal("lou", summary["user"])
	})

	It("shows config values", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"gitlab.com\"\n"), 0o644)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		output, err := testhelpers.ExecuteCmd(rootCmd, "config", "show")

		assert.NoError(err)
		var payload map[string]any
		err = json.Unmarshal([]byte(output), &payload)
		assert.NoError(err)
		assert.Equal(configPath, payload["path"])
		assert.Equal("gitlab.com", payload["site"])
		assert.Equal("lou", payload["user"])
	})

	It("removes the config file", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")
		err := os.WriteFile(configPath, []byte("user = \"lou\"\n"), 0o644)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		_, err = testhelpers.ExecuteCmd(rootCmd, "config", "remove")
		assert.NoError(err)

		_, err = os.Stat(configPath)
		assert.ErrorIs(err, os.ErrNotExist)
	})

	It("uses a repo-local gtk-config.toml when present", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "gtk-config.toml")

		currentDir, err := os.Getwd()
		assert.NoError(err)

		err = os.Chdir(tempDir)
		assert.NoError(err)
		defer func() {
			_ = os.Chdir(currentDir)
		}()

		err = os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"github.com\"\n"), 0o644)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
		})

		output, err := testhelpers.ExecuteCmd(rootCmd, "config", "show")

		assert.NoError(err)
		var payload map[string]any
		err = json.Unmarshal([]byte(output), &payload)
		assert.NoError(err)
		assert.Equal(configPath, payload["path"])
	})

	It("adds a provider mapping", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		_, err := testhelpers.ExecuteCmd(rootCmd, "config", "provider", "add", "--name", "gitlab", "--path", "/tmp/gitlab")

		assert.NoError(err)

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Len(values.Providers, 1)
		assert.Equal("gitlab", values.Providers[0].Name)
		assert.Equal("/tmp/gitlab", values.Providers[0].Path)
	})

	It("removes a provider mapping", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("[[providers]]\nname = \"gitlab\"\npath = \"/tmp/gitlab\"\n"), 0o644)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		_, err = testhelpers.ExecuteCmd(rootCmd, "config", "provider", "remove", "--name", "gitlab")

		assert.NoError(err)

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Empty(values.Providers)
	})

	It("lists provider mappings", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("[[providers]]\nname = \"gitlab\"\npath = \"/tmp/gitlab\"\n"), 0o644)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		output, err := testhelpers.ExecuteCmd(rootCmd, "config", "provider", "list")

		assert.NoError(err)
		assert.Contains(output, "gitlab")
		assert.Contains(output, "/tmp/gitlab")
	})

	It("adds and shows package presets", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		_, err := testhelpers.ExecuteCmd(
			rootCmd,
			"config",
			"package-preset",
			"add",
			"--name",
			"cli",
			"--package",
			"github.com/spf13/cobra",
			"--package",
			"github.com/spf13/viper",
		)
		assert.NoError(err)

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Len(values.PackagePresets, 1)
		assert.Equal([]string{"github.com/spf13/cobra", "github.com/spf13/viper"}, values.PackagePresets["cli"])

		output, err := testhelpers.ExecuteCmd(rootCmd, "config", "package-preset", "list")
		assert.NoError(err)
		assert.Contains(output, "cli")
		assert.Contains(output, "github.com/spf13/cobra")
	})

	It("rejects empty package preset keys during flag validation", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		_, err := testhelpers.ExecuteCmd(
			rootCmd,
			"config",
			"package-preset",
			"add",
			"--name",
			"   ",
			"--package",
			"github.com/spf13/cobra",
		)

		assert.Error(err)
		assert.Contains(err.Error(), "invalid argument")
		assert.Contains(err.Error(), "name must not be empty")
	})

	It("updates provider assurance", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		_, err := testhelpers.ExecuteCmd(rootCmd, "config", "set-assure-providers", "true")
		assert.NoError(err)

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.True(values.AssureProviders)
	})
})
