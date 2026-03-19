package cmd_test

import (
	"os"
	"path/filepath"

	"github.com/louiss0/go-toolkit/cmd"
	"github.com/louiss0/go-toolkit/internal/testhelpers"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var Scaffold = Describe("scaffold command", func() {
	assert := assert.New(GinkgoT())

	It("scaffolds a folder with a README", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		target := filepath.Join(tempDir, "demo")

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   filepath.Join(tempDir, "config.toml"),
		})

		_, err := testhelpers.ExecuteCmd(rootCmd, "scaffold", "demo", "--folder", target, "--readme")

		assert.NoError(err)
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		content, err := os.ReadFile(filepath.Join(target, "README.md"))
		assert.NoError(err)
		assert.Contains(string(content), "# demo")
	})

	It("scaffolds and initializes a module", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		target := filepath.Join(tempDir, "demo")
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"github.com\"\n"), 0o644)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		runner.On("Run", mock.Anything, "go", []string{"-C", target, "mod", "init", "github.com/lou/demo"}).Return(nil).Once()

		_, err = testhelpers.ExecuteCmd(rootCmd, "scaffold", "demo", "--folder", target, "--module")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})
})
