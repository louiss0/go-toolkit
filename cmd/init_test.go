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
	"github.com/stretchr/testify/mock"
)

var Init = Describe("init command", func() {
	assert := assert.New(GinkgoT())

	It("inits a module using the registered user", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")
		workingDir, err := os.Getwd()
		assert.NoError(err)

		err = os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"github.com\"\n"), 0o644)
		assert.NoError(err)

		err = os.Chdir(tempDir)
		assert.NoError(err)
		DeferCleanup(func() {
			_ = os.Chdir(workingDir)
		})

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		runner.On("Run", mock.Anything, "go", []string{"mod", "init", "github.com/lou/toolkit"}).Return(nil).Once()
		runner.On("Run", mock.Anything, "git", []string{"init"}).Return(nil).Once()

		_, err = testhelpers.ExecuteCmd(rootCmd, "init", "toolkit")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())

		_, err = os.Stat(filepath.Join(tempDir, "internal"))
		assert.NoError(err)

		content, err := os.ReadFile(filepath.Join(tempDir, "main.go"))
		assert.NoError(err)
		assert.Contains(string(content), "package main")
	})

	It("prompts for init details when no args are provided", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")
		workingDir, err := os.Getwd()
		assert.NoError(err)

		err = os.Chdir(tempDir)
		assert.NoError(err)
		DeferCleanup(func() {
			_ = os.Chdir(workingDir)
		})

		promptRunner := testhelpers.NewPromptRunnerMock(
			testhelpers.PromptStep{Kind: testhelpers.PromptStepInput, Value: "toolkit"},
			testhelpers.PromptStep{Kind: testhelpers.PromptStepInput, Value: "lou"},
			testhelpers.PromptStep{Kind: testhelpers.PromptStepSelect, Value: "github.com"},
			testhelpers.PromptStep{Kind: testhelpers.PromptStepSelect, Value: "library"},
			testhelpers.PromptStep{Kind: testhelpers.PromptStepSelect, Value: "yes"},
			testhelpers.PromptStep{Kind: testhelpers.PromptStepSelect, Value: "no"},
			testhelpers.PromptStep{Kind: testhelpers.PromptStepInput, Value: "github.com/samber/lo"},
		)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: promptRunner,
			ConfigPath:   configPath,
		})

		runner.On("Run", mock.Anything, "go", []string{"mod", "init", "github.com/lou/toolkit"}).Return(nil).Once()
		runner.On("Run", mock.Anything, "go", []string{"get", "github.com/samber/lo"}).Return(nil).Once()

		output, err := testhelpers.ExecuteCmd(rootCmd, "init")

		assert.NoError(err)
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, "git", mock.Anything)

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Equal("lou", values.User)
		assert.Equal("github.com", values.Site)
		assert.True(values.Scaffold.WriteTests)

		_, err = os.Stat(filepath.Join(tempDir, "main.go"))
		assert.Error(err)
		_, err = os.Stat(filepath.Join(tempDir, "internal"))
		assert.NoError(err)

		var summary map[string]any
		err = json.Unmarshal([]byte(output), &summary)
		assert.NoError(err)
		assert.Equal("github.com/lou/toolkit", summary["module_path"])
	})
})
