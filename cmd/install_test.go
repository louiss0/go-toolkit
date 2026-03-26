package cmd_test

import (
	"os"
	"path/filepath"

	"github.com/louiss0/go-toolkit/cmd"
	"github.com/louiss0/go-toolkit/internal/modindex/config"
	"github.com/louiss0/go-toolkit/internal/testhelpers"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var Install = Describe("install command", func() {
	assert := assert.New(GinkgoT())

	It("installs a package globally and saves it to config", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := writeDefaultConfig(configPath)
		assert.NoError(err)

		runner.On("Run", mock.Anything, "go", []string{"install", "github.com/onsi/ginkgo/v2@latest"}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		output, err := testhelpers.ExecuteCmd(rootCmd, "install", "github.com/onsi/ginkgo/v2")

		assert.NoError(err)
		assert.Contains(output, "installed and saved to global packages")
		runner.AssertExpectations(GinkgoT())

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Contains(values.GlobalPackages, "github.com/onsi/ginkgo/v2")
	})

	It("installs multiple packages globally", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := writeDefaultConfig(configPath)
		assert.NoError(err)

		runner.On("Run", mock.Anything, "go", []string{"install", "github.com/onsi/ginkgo/v2@latest"}).Return(nil).Once()
		runner.On("Run", mock.Anything, "go", []string{"install", "github.com/carapace-sh/carapace@latest"}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		_, err = testhelpers.ExecuteCmd(rootCmd, "install", "github.com/onsi/ginkgo/v2", "github.com/carapace-sh/carapace")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Contains(values.GlobalPackages, "github.com/onsi/ginkgo/v2")
		assert.Contains(values.GlobalPackages, "github.com/carapace-sh/carapace")
	})

	It("installs a short package path with a major version suffix globally", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := writeDefaultConfig(configPath)
		assert.NoError(err)

		runner.On("Run", mock.Anything, "go", []string{"install", "github.com/onsi/ginkgo/v2@latest"}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		_, err = testhelpers.ExecuteCmd(rootCmd, "install", "onsi/ginkgo/v2")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Contains(values.GlobalPackages, "github.com/onsi/ginkgo/v2")
	})

	It("prints the install command on dry run", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := writeDefaultConfig(configPath)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		output, err := testhelpers.ExecuteCmd(rootCmd, "install", "github.com/onsi/ginkgo/v2", "--dry-run")

		assert.NoError(err)
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything)
		assert.Contains(output, "go install github.com/onsi/ginkgo/v2@latest")
	})

	It("does not duplicate global packages on repeated install", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"github.com\"\nglobal_packages = [\"github.com/onsi/ginkgo/v2\"]\n"), 0o644)
		assert.NoError(err)

		runner.On("Run", mock.Anything, "go", []string{"install", "github.com/onsi/ginkgo/v2@latest"}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		_, err = testhelpers.ExecuteCmd(rootCmd, "install", "github.com/onsi/ginkgo/v2")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Len(values.GlobalPackages, 1)
	})

	It("installs packages from presets", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"github.com\"\n[package_presets]\ncli = [\"onsi/ginkgo\", \"carapace-sh/carapace\"]\n"), 0o644)
		assert.NoError(err)

		runner.On("Run", mock.Anything, "go", []string{"install", "github.com/onsi/ginkgo@latest"}).Return(nil).Once()
		runner.On("Run", mock.Anything, "go", []string{"install", "github.com/carapace-sh/carapace@latest"}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		_, err = testhelpers.ExecuteCmd(rootCmd, "install", "--preset", "cli")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("rejects @none for install", func() {
		configPath := ""

		installCmd := cmd.NewInstallCmd(&testhelpers.RunnerMock{}, testhelpers.NewPromptRunnerMock(), &configPath)

		err := installCmd.Args(installCmd, []string{"github.com/onsi/ginkgo@none"})

		assert.Error(err)
		assert.Contains(err.Error(), "@none is not valid")
	})

	It("rejects @none in --package values", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := writeDefaultConfig(configPath)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		_, err = testhelpers.ExecuteCmd(rootCmd, "install", "--package", "github.com/onsi/ginkgo/v2@none")

		assert.Error(err)
		assert.Contains(err.Error(), "@none is not valid")
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything)
	})

	It("rejects @none in preset package values", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"github.com\"\n[package_presets]\ncli = [\"github.com/onsi/ginkgo/v2@none\"]\n"), 0o644)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		_, err = testhelpers.ExecuteCmd(rootCmd, "install", "--preset", "cli")

		assert.Error(err)
		assert.Contains(err.Error(), "@none is not valid")
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything)
	})
})
