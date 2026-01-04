package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/louiss0/cobra-cli-template/cmd"
	"github.com/louiss0/cobra-cli-template/internal/config"
	"github.com/louiss0/cobra-cli-template/internal/prompt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type runnerMock struct {
	mock.Mock
}

func (m *runnerMock) Run(cmd *cobra.Command, name string, args ...string) error {
	call := m.Called(cmd, name, args)
	return call.Error(0)
}

func ExecuteCmd(cmd *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	errBuff := new(bytes.Buffer)

	cmd.SetOut(buf)
	cmd.SetErr(errBuff)
	cmd.SetArgs(args)

	err := cmd.Execute()
	if errBuff.Len() > 0 {
		return "", fmt.Errorf("command failed: %s", errBuff.String())
	}

	return buf.String(), err
}

type promptStepKind string

const (
	promptStepInput  promptStepKind = "input"
	promptStepSelect promptStepKind = "select"
)

type promptStep struct {
	kind  promptStepKind
	value string
	err   error
}

type promptRunnerMock struct {
	steps []promptStep
	index int
}

func newPromptRunnerMock(steps ...promptStep) *promptRunnerMock {
	return &promptRunnerMock{steps: steps}
}

func (m *promptRunnerMock) Input(_ *cobra.Command, input prompt.Input) (string, error) {
	step, err := m.next(promptStepInput)
	if err != nil {
		return "", err
	}
	if step.err != nil {
		return "", step.err
	}
	if input.Validate != nil {
		if err := input.Validate(step.value); err != nil {
			return "", err
		}
	}
	return step.value, nil
}

func (m *promptRunnerMock) Select(_ *cobra.Command, selectInput prompt.Select) (string, error) {
	step, err := m.next(promptStepSelect)
	if err != nil {
		return "", err
	}
	if step.err != nil {
		return "", step.err
	}
	for _, option := range selectInput.Options {
		if option.Value == step.value {
			return step.value, nil
		}
	}
	return "", fmt.Errorf("unexpected selection: %s", step.value)
}

func (m *promptRunnerMock) next(expected promptStepKind) (promptStep, error) {
	if m.index >= len(m.steps) {
		return promptStep{}, fmt.Errorf("prompt mock: no steps remaining")
	}
	step := m.steps[m.index]
	m.index++
	if step.kind != expected {
		return promptStep{}, fmt.Errorf("prompt mock: expected %s, got %s", expected, step.kind)
	}
	return step, nil
}

var Test = Describe("test command", func() {
	assert := assert.New(GinkgoT())

	It("runs go test for all packages by default", func() {
		runner := &runnerMock{}
		runner.On("Run", mock.Anything, "go", []string{"test", "./..."}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: runner,
		})

		_, err := ExecuteCmd(rootCmd, "test")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})
})

var RootOptions = Describe("root options", func() {
	assert := assert.New(GinkgoT())

	It("panics when the config path is whitespace", func() {
		assert.Panics(func() {
			_ = cmd.NewRootCmdWithOptions(cmd.RootOptions{
				ConfigPath: "   ",
			})
		})
	})
})

var Remove = Describe("remove command", func() {
	assert := assert.New(GinkgoT())

	It("removes a fully qualified module", func() {
		runner := &runnerMock{}
		runner.On("Run", mock.Anything, "go", []string{"get", "github.com/acme/tool@none"}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: runner,
		})

		_, err := ExecuteCmd(rootCmd, "remove", "github.com/acme/tool")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("removes multiple modules in one command", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"github.com\"\n"), 0o644)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     runner,
			ConfigPath: configPath,
		})

		runner.On(
			"Run",
			mock.Anything,
			"go",
			[]string{"get", "github.com/lou/tool@none", "github.com/acme/other@none"},
		).Return(nil).Once()

		_, err = ExecuteCmd(rootCmd, "remove", "tool", "github.com/acme/other")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("removes a module path with a version suffix", func() {
		runner := &runnerMock{}
		runner.On("Run", mock.Anything, "go", []string{"get", "github.com/onsi/ginkgo/v2@none"}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: runner,
		})

		_, err := ExecuteCmd(rootCmd, "remove", "github.com/onsi/ginkgo/v2")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("rejects @none for remove input", func() {
		configPath := ""

		removeCmd := cmd.NewRemoveCmd(&runnerMock{}, &configPath)

		err := removeCmd.Args(removeCmd, []string{"github.com/onsi/ginkgo@none"})

		assert.Error(err)
		assert.Contains(err.Error(), "added automatically")
	})

	It("prints the remove command on dry run", func() {
		runner := &runnerMock{}

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: runner,
		})

		output, err := ExecuteCmd(rootCmd, "remove", "github.com/onsi/ginkgo/v2", "--dry-run")

		assert.NoError(err)
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		assert.Contains(output, "go get github.com/onsi/ginkgo/v2@none")
	})
})

var Init = Describe("init command", func() {
	assert := assert.New(GinkgoT())

	It("inits a module using the registered user", func() {
		runner := &runnerMock{}
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
			Runner:     runner,
			ConfigPath: configPath,
		})

		runner.On("Run", mock.Anything, "go", []string{"mod", "init", "github.com/lou/toolkit"}).Return(nil).Once()
		runner.On("Run", mock.Anything, "git", []string{"init"}).Return(nil).Once()

		_, err = ExecuteCmd(rootCmd, "init", "toolkit")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())

		_, err = os.Stat(filepath.Join(tempDir, "internal"))
		assert.NoError(err)

		content, err := os.ReadFile(filepath.Join(tempDir, "main.go"))
		assert.NoError(err)
		assert.Contains(string(content), "package main")
	})

	It("prompts for init details when no args are provided", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")
		workingDir, err := os.Getwd()
		assert.NoError(err)

		err = os.Chdir(tempDir)
		assert.NoError(err)
		DeferCleanup(func() {
			_ = os.Chdir(workingDir)
		})

		promptRunner := newPromptRunnerMock(
			promptStep{kind: promptStepInput, value: "toolkit"},
			promptStep{kind: promptStepInput, value: "lou"},
			promptStep{kind: promptStepSelect, value: "github.com"},
			promptStep{kind: promptStepSelect, value: "library"},
			promptStep{kind: promptStepSelect, value: "yes"},
			promptStep{kind: promptStepSelect, value: "no"},
			promptStep{kind: promptStepInput, value: "github.com/samber/lo"},
		)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: promptRunner,
			ConfigPath:   configPath,
		})

		runner.On("Run", mock.Anything, "go", []string{"mod", "init", "github.com/lou/toolkit"}).Return(nil).Once()
		runner.On("Run", mock.Anything, "go", []string{"get", "github.com/samber/lo"}).Return(nil).Once()

		output, err := ExecuteCmd(rootCmd, "init")

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

var Scaffold = Describe("scaffold command", func() {
	assert := assert.New(GinkgoT())

	It("scaffolds a folder with a README", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		target := filepath.Join(tempDir, "demo")

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     runner,
			ConfigPath: filepath.Join(tempDir, "config.toml"),
		})

		_, err := ExecuteCmd(rootCmd, "scaffold", "demo", "--folder", target, "--readme")

		assert.NoError(err)
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		content, err := os.ReadFile(filepath.Join(target, "README.md"))
		assert.NoError(err)
		assert.Contains(string(content), "# demo")
	})

	It("scaffolds and initializes a module", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		target := filepath.Join(tempDir, "demo")
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"github.com\"\n"), 0o644)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     runner,
			ConfigPath: configPath,
		})

		runner.On("Run", mock.Anything, "go", []string{"-C", target, "mod", "init", "github.com/lou/demo"}).Return(nil).Once()

		_, err = ExecuteCmd(rootCmd, "scaffold", "demo", "--folder", target, "--module")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})
})

var Add = Describe("add command", func() {
	assert := assert.New(GinkgoT())

	It("adds multiple packages with short paths", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"github.com\"\n"), 0o644)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     runner,
			ConfigPath: configPath,
		})

		runner.On(
			"Run",
			mock.Anything,
			"go",
			[]string{"get", "github.com/samber/lo", "github.com/stretchr/testify", "github.com/onsi/ginkgo"},
		).Return(nil).Once()

		_, err = ExecuteCmd(rootCmd, "add", "samber/lo", "stretchr/testify", "onsi/ginkgo")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("adds a module path with a version suffix", func() {
		runner := &runnerMock{}
		runner.On("Run", mock.Anything, "go", []string{"get", "github.com/onsi/ginkgo/v2"}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: runner,
		})

		_, err := ExecuteCmd(rootCmd, "add", "github.com/onsi/ginkgo/v2")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("adds a module with an @version suffix", func() {
		runner := &runnerMock{}
		runner.On("Run", mock.Anything, "go", []string{"get", "github.com/onsi/ginkgo@v2.0.0"}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: runner,
		})

		_, err := ExecuteCmd(rootCmd, "add", "github.com/onsi/ginkgo@v2.0.0")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("rejects @none for add", func() {
		configPath := ""

		addCmd := cmd.NewAddCmd(&runnerMock{}, &configPath)

		err := addCmd.Args(addCmd, []string{"github.com/onsi/ginkgo@none"})

		assert.Error(err)
		assert.Contains(err.Error(), "use remove")
	})

	It("prints the add command on dry run", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"github.com\"\n"), 0o644)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     runner,
			ConfigPath: configPath,
		})

		output, err := ExecuteCmd(rootCmd, "add", "samber/lo", "--dry-run")

		assert.NoError(err)
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		assert.Contains(output, "go get github.com/samber/lo")
	})
})

var Config = Describe("config command", func() {
	assert := assert.New(GinkgoT())

	It("initializes config with defaults", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     runner,
			ConfigPath: configPath,
		})

		_, err := ExecuteCmd(rootCmd, "config", "init", "--user", "lou")

		assert.NoError(err)

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Equal("lou", values.User)
		assert.Equal("github.com", values.Site)
	})

	It("prompts for config init when no flags are provided", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		promptRunner := newPromptRunnerMock(
			promptStep{kind: promptStepInput, value: "lou"},
			promptStep{kind: promptStepSelect, value: "gitlab.com"},
		)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: promptRunner,
			ConfigPath:   configPath,
		})

		output, err := ExecuteCmd(rootCmd, "config", "init")

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
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"gitlab.com\"\n"), 0o644)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     runner,
			ConfigPath: configPath,
		})

		output, err := ExecuteCmd(rootCmd, "config", "show")

		assert.NoError(err)
		var payload map[string]any
		err = json.Unmarshal([]byte(output), &payload)
		assert.NoError(err)
		assert.Equal(configPath, payload["path"])
		assert.Equal("gitlab.com", payload["site"])
		assert.Equal("lou", payload["user"])
	})

	It("uses a repo-local gtk-config.toml when present", func() {
		runner := &runnerMock{}
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
			Runner: runner,
		})

		output, err := ExecuteCmd(rootCmd, "config", "show")

		assert.NoError(err)
		var payload map[string]any
		err = json.Unmarshal([]byte(output), &payload)
		assert.NoError(err)
		assert.Equal(configPath, payload["path"])
	})

	It("adds a provider mapping", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     runner,
			ConfigPath: configPath,
		})

		_, err := ExecuteCmd(rootCmd, "config", "provider", "add", "--name", "gitlab", "--path", "/tmp/gitlab")

		assert.NoError(err)

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Len(values.Providers, 1)
		assert.Equal("gitlab", values.Providers[0].Name)
		assert.Equal("/tmp/gitlab", values.Providers[0].Path)
	})

	It("removes a provider mapping", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("[[providers]]\nname = \"gitlab\"\npath = \"/tmp/gitlab\"\n"), 0o644)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     runner,
			ConfigPath: configPath,
		})

		_, err = ExecuteCmd(rootCmd, "config", "provider", "remove", "--name", "gitlab")

		assert.NoError(err)

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Empty(values.Providers)
	})

	It("lists provider mappings", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("[[providers]]\nname = \"gitlab\"\npath = \"/tmp/gitlab\"\n"), 0o644)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     runner,
			ConfigPath: configPath,
		})

		output, err := ExecuteCmd(rootCmd, "config", "provider", "list")

		assert.NoError(err)
		assert.Contains(output, "gitlab")
		assert.Contains(output, "/tmp/gitlab")
	})
})

var Search = Describe("search command", func() {
	assert := assert.New(GinkgoT())

	It("accepts a scope and package query", func() {
		searchCmd := cmd.NewSearchCmd()

		err := searchCmd.Args(searchCmd, []string{"acme/tool"})

		assert.NoError(err)
	})

	It("rejects missing query input", func() {
		searchCmd := cmd.NewSearchCmd()

		err := searchCmd.Args(searchCmd, []string{})

		assert.Error(err)
		assert.Contains(err.Error(), "accepts 1 arg(s)")
	})

	It("rejects queries without a scope", func() {
		searchCmd := cmd.NewSearchCmd()

		err := searchCmd.Args(searchCmd, []string{"tool"})

		assert.Error(err)
		assert.Contains(err.Error(), "scope/package")
	})

	It("accepts queries with extra path segments", func() {
		searchCmd := cmd.NewSearchCmd()

		err := searchCmd.Args(searchCmd, []string{"github.com/acme/tool"})

		assert.NoError(err)
	})

	It("accepts prefixed queries with multiple segments", func() {
		searchCmd := cmd.NewSearchCmd()

		err := searchCmd.Args(searchCmd, []string{"github.com/acme/tool/extra"})

		assert.NoError(err)
	})
})
