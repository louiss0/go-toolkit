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

var Remove = Describe("remove command", func() {
	assert := assert.New(GinkgoT())

	It("removes a fully qualified module", func() {
		runner := &testhelpers.RunnerMock{}
		runner.On("Run", mock.Anything, "go", []string{"get", "github.com/acme/tool@none"}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
		})

		_, err := testhelpers.ExecuteCmd(rootCmd, "remove", "github.com/acme/tool")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("removes multiple modules in one command", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"github.com\"\n"), 0o644)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		runner.On(
			"Run",
			mock.Anything,
			"go",
			[]string{"get", "github.com/lou/tool@none", "github.com/acme/other@none"},
		).Return(nil).Once()

		_, err = testhelpers.ExecuteCmd(rootCmd, "remove", "tool", "github.com/acme/other")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("removes a module path with a version suffix", func() {
		runner := &testhelpers.RunnerMock{}
		runner.On("Run", mock.Anything, "go", []string{"get", "github.com/onsi/ginkgo/v2@none"}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
		})

		_, err := testhelpers.ExecuteCmd(rootCmd, "remove", "github.com/onsi/ginkgo/v2")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("rejects @none for remove input", func() {
		configPath := ""

		removeCmd := cmd.NewRemoveCmd(&testhelpers.RunnerMock{}, &configPath)

		err := removeCmd.Args(removeCmd, []string{"github.com/onsi/ginkgo@none"})

		assert.Error(err)
		assert.Contains(err.Error(), "added automatically")
	})

	It("prints the remove command on dry run", func() {
		runner := &testhelpers.RunnerMock{}

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
		})

		output, err := testhelpers.ExecuteCmd(rootCmd, "remove", "github.com/onsi/ginkgo/v2", "--dry-run")

		assert.NoError(err)
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		assert.Contains(output, "go get github.com/onsi/ginkgo/v2@none")
	})
})
