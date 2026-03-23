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

var Add = Describe("add command", func() {
	assert := assert.New(GinkgoT())

	It("adds multiple packages with short paths", func() {
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
			[]string{"get", "github.com/samber/lo", "github.com/stretchr/testify", "github.com/onsi/ginkgo"},
		).Return(nil).Once()

		_, err = testhelpers.ExecuteCmd(rootCmd, "add", "samber/lo", "stretchr/testify", "onsi/ginkgo")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("adds a module path with a version suffix", func() {
		runner := &testhelpers.RunnerMock{}
		runner.On("Run", mock.Anything, "go", []string{"get", "github.com/onsi/ginkgo/v2"}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
		})

		_, err := testhelpers.ExecuteCmd(rootCmd, "add", "github.com/onsi/ginkgo/v2")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("adds a module with an @version suffix", func() {
		runner := &testhelpers.RunnerMock{}
		runner.On("Run", mock.Anything, "go", []string{"get", "github.com/onsi/ginkgo@v2.0.0"}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
		})

		_, err := testhelpers.ExecuteCmd(rootCmd, "add", "github.com/onsi/ginkgo@v2.0.0")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("rejects @none for add", func() {
		configPath := ""

		addCmd := cmd.NewAddCmd(&testhelpers.RunnerMock{}, testhelpers.NewPromptRunnerMock(), &configPath)

		err := addCmd.Args(addCmd, []string{"github.com/onsi/ginkgo@none"})

		assert.Error(err)
		assert.Contains(err.Error(), "use remove")
	})

	It("prints the add command on dry run", func() {
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

		output, err := testhelpers.ExecuteCmd(rootCmd, "add", "samber/lo", "--dry-run")

		assert.NoError(err)
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		assert.Contains(output, "go get github.com/samber/lo")
	})

	It("prompts for packages when no args are provided", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"github.com\"\n"), 0o644)
		assert.NoError(err)

		promptRunner := testhelpers.NewPromptRunnerMock(
			testhelpers.PromptStep{Kind: testhelpers.PromptStepInput, Value: "samber/lo, stretchr/testify"},
		)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: promptRunner,
			ConfigPath:   configPath,
		})

		runner.On(
			"Run",
			mock.Anything,
			"go",
			[]string{"get", "github.com/samber/lo", "github.com/stretchr/testify"},
		).Return(nil).Once()

		_, err = testhelpers.ExecuteCmd(rootCmd, "add")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("adds packages from presets", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"github.com\"\n[package_presets]\ncli = [\"samber/lo\", \"stretchr/testify\"]\n"), 0o644)
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
			[]string{"get", "github.com/samber/lo", "github.com/stretchr/testify"},
		).Return(nil).Once()

		_, err = testhelpers.ExecuteCmd(rootCmd, "add", "--preset", "cli")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})
})
