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

var Uninstall = Describe("uninstall command", func() {
	assert := assert.New(GinkgoT())

	It("uninstalls a package globally and removes it from config", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"github.com\"\nglobal_packages = [\"github.com/onsi/ginkgo/v2\"]\n"), 0o644)
		assert.NoError(err)

		runner.On("Run", mock.Anything, "go", []string{"clean", "-i", "github.com/onsi/ginkgo/v2"}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		output, err := testhelpers.ExecuteCmd(rootCmd, "uninstall", "github.com/onsi/ginkgo/v2")

		assert.NoError(err)
		assert.Contains(output, "uninstalled and removed from global packages")
		runner.AssertExpectations(GinkgoT())

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Empty(values.GlobalPackages)
	})

	It("prints the uninstall command on dry run", func() {
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

		output, err := testhelpers.ExecuteCmd(rootCmd, "uninstall", "github.com/onsi/ginkgo/v2", "--dry-run")

		assert.NoError(err)
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything)
		assert.Contains(output, "go clean -i github.com/onsi/ginkgo/v2")
	})
})
