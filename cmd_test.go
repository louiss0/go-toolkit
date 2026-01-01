package main_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/louiss0/cobra-cli-template/cmd"
	"github.com/louiss0/cobra-cli-template/internal/config"
	"github.com/louiss0/cobra-cli-template/internal/modindex"
	. "github.com/onsi/ginkgo/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type runnerMock struct {
	mock.Mock
}

func (m *runnerMock) Run(name string, args []string, stdout, stderr io.Writer) error {
	call := m.Called(name, args, stdout, stderr)
	return call.Error(0)
}

type indexFetcherMock struct {
	mock.Mock
}

func (m *indexFetcherMock) Fetch(ctx context.Context, request modindex.Request) ([]modindex.Entry, error) {
	call := m.Called(ctx, request)
	entries, _ := call.Get(0).([]modindex.Entry)
	return entries, call.Error(1)
}

func executeCmd(cmd *cobra.Command, args ...string) (string, error) {
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

var _ = Describe("CLI", func() {
	assert := assert.New(GinkgoT())

	It("runs go test for all packages by default", func() {
		runner := &runnerMock{}
		runner.On("Run", "go", []string{"test", "./..."}, mock.Anything, mock.Anything).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: runner,
		})

		_, err := executeCmd(rootCmd, "test")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("removes a fully qualified module", func() {
		runner := &runnerMock{}
		runner.On("Run", "go", []string{"get", "github.com/acme/tool@none"}, mock.Anything, mock.Anything).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: runner,
		})

		_, err := executeCmd(rootCmd, "remove", "github.com/acme/tool")

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
			"go",
			[]string{"get", "github.com/lou/tool@none", "github.com/acme/other@none"},
			mock.Anything,
			mock.Anything,
		).Return(nil).Once()

		_, err = executeCmd(rootCmd, "remove", "tool", "github.com/acme/other")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("removes a module path with a version suffix", func() {
		runner := &runnerMock{}
		runner.On("Run", "go", []string{"get", "github.com/onsi/ginkgo/v2@none"}, mock.Anything, mock.Anything).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: runner,
		})

		_, err := executeCmd(rootCmd, "remove", "github.com/onsi/ginkgo/v2")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("rejects @none for remove input", func() {
		runner := &runnerMock{}

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: runner,
		})

		_, err := executeCmd(rootCmd, "remove", "github.com/onsi/ginkgo@none")

		assert.Error(err)
		assert.Contains(err.Error(), "added automatically")
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	It("prints the remove command on dry run", func() {
		runner := &runnerMock{}

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: runner,
		})

		output, err := executeCmd(rootCmd, "remove", "github.com/onsi/ginkgo/v2", "--dry-run")

		assert.NoError(err)
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		assert.Contains(output, "go get github.com/onsi/ginkgo/v2@none")
	})

	It("inits a module using the registered user", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"github.com\"\n"), 0o644)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     runner,
			ConfigPath: configPath,
		})

		runner.On("Run", "go", []string{"mod", "init", "github.com/lou/toolkit"}, mock.Anything, mock.Anything).Return(nil).Once()

		_, err = executeCmd(rootCmd, "init", "toolkit")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("searches the module index with a default site filter", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")
		indexFetcher := &indexFetcherMock{}
		entries := []modindex.Entry{
			{Path: "github.com/acme/tool", Version: "v1.0.0"},
			{Path: "gitlab.com/acme/tool", Version: "v1.2.0"},
		}
		indexFetcher.On("Fetch", mock.Anything, modindex.Request{Limit: 200}).Return(entries, nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			IndexFetcher: indexFetcher,
			ConfigPath:   configPath,
		})

		output, err := executeCmd(rootCmd, "search", "tool")

		assert.NoError(err)
		assert.Contains(output, "github.com/acme/tool")
		assert.NotContains(output, "gitlab.com/acme/tool")
		indexFetcher.AssertExpectations(GinkgoT())
	})

	It("searches without a site filter when the query includes a domain", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")
		indexFetcher := &indexFetcherMock{}
		entries := []modindex.Entry{
			{Path: "github.com/acme/tool", Version: "v1.0.0"},
			{Path: "gitlab.com/acme/tool", Version: "v1.2.0"},
		}
		indexFetcher.On("Fetch", mock.Anything, modindex.Request{Limit: 200}).Return(entries, nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			IndexFetcher: indexFetcher,
			ConfigPath:   configPath,
		})

		output, err := executeCmd(rootCmd, "search", "gitlab.com")

		assert.NoError(err)
		assert.Contains(output, "gitlab.com/acme/tool")
		assert.NotContains(output, "github.com/acme/tool")
		indexFetcher.AssertExpectations(GinkgoT())
	})

	It("prints details when requested", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")
		indexFetcher := &indexFetcherMock{}
		entries := []modindex.Entry{
			{Path: "github.com/acme/tool", Version: "v1.2.3", Timestamp: "2024-01-01T00:00:00Z"},
		}
		indexFetcher.On("Fetch", mock.Anything, modindex.Request{Limit: 200}).Return(entries, nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			IndexFetcher: indexFetcher,
			ConfigPath:   configPath,
		})

		output, err := executeCmd(rootCmd, "search", "tool", "--details")

		assert.NoError(err)
		assert.Contains(output, "github.com/acme/tool")
		assert.Contains(output, "v1.2.3")
		assert.Contains(output, "2024-01-01T00:00:00Z")
		indexFetcher.AssertExpectations(GinkgoT())
	})

	It("rejects unknown sites without --full", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")
		indexFetcher := &indexFetcherMock{}

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			IndexFetcher: indexFetcher,
			ConfigPath:   configPath,
		})

		_, err := executeCmd(rootCmd, "search", "tool", "--site", "example.com")

		assert.Error(err)
		assert.Contains(err.Error(), "unsupported site")
		indexFetcher.AssertNotCalled(GinkgoT(), "Fetch", mock.Anything, mock.Anything)
	})

	It("rejects sites without a dot", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")
		indexFetcher := &indexFetcherMock{}

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			IndexFetcher: indexFetcher,
			ConfigPath:   configPath,
		})

		_, err := executeCmd(rootCmd, "search", "tool", "--site", "githubcom")

		assert.Error(err)
		assert.Contains(err.Error(), "sitename.domain")
		indexFetcher.AssertNotCalled(GinkgoT(), "Fetch", mock.Anything, mock.Anything)
	})

	It("scaffolds a folder with a README", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		target := filepath.Join(tempDir, "demo")

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     runner,
			ConfigPath: filepath.Join(tempDir, "config.toml"),
		})

		_, err := executeCmd(rootCmd, "scaffold", "demo", "--folder", target, "--readme")

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

		runner.On("Run", "go", []string{"-C", target, "mod", "init", "github.com/lou/demo"}, mock.Anything, mock.Anything).Return(nil).Once()

		_, err = executeCmd(rootCmd, "scaffold", "demo", "--folder", target, "--module")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

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
			"go",
			[]string{"get", "github.com/samber/lo", "github.com/stretchr/testify", "github.com/onsi/ginkgo"},
			mock.Anything,
			mock.Anything,
		).Return(nil).Once()

		_, err = executeCmd(rootCmd, "add", "samber/lo", "stretchr/testify", "onsi/ginkgo")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("adds a module path with a version suffix", func() {
		runner := &runnerMock{}
		runner.On("Run", "go", []string{"get", "github.com/onsi/ginkgo/v2"}, mock.Anything, mock.Anything).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: runner,
		})

		_, err := executeCmd(rootCmd, "add", "github.com/onsi/ginkgo/v2")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("adds a module with an @version suffix", func() {
		runner := &runnerMock{}
		runner.On("Run", "go", []string{"get", "github.com/onsi/ginkgo@v2.0.0"}, mock.Anything, mock.Anything).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: runner,
		})

		_, err := executeCmd(rootCmd, "add", "github.com/onsi/ginkgo@v2.0.0")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})

	It("rejects @none for add", func() {
		runner := &runnerMock{}

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: runner,
		})

		_, err := executeCmd(rootCmd, "add", "github.com/onsi/ginkgo@none")

		assert.Error(err)
		assert.Contains(err.Error(), "use remove")
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
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

		output, err := executeCmd(rootCmd, "add", "samber/lo", "--dry-run")

		assert.NoError(err)
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		assert.Contains(output, "go get github.com/samber/lo")
	})

	It("initializes config with defaults", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     runner,
			ConfigPath: configPath,
		})

		_, err := executeCmd(rootCmd, "config", "init", "--user", "lou")

		assert.NoError(err)

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Equal("lou", values.User)
		assert.Equal("github.com", values.Site)
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

		output, err := executeCmd(rootCmd, "config", "show")

		assert.NoError(err)
		assert.Contains(output, "path:")
		assert.Contains(output, "site: gitlab.com")
		assert.Contains(output, "user: lou")
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

		output, err := executeCmd(rootCmd, "config", "show")

		assert.NoError(err)
		assert.Contains(output, "path: "+configPath)
	})

	It("adds a provider mapping", func() {
		runner := &runnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     runner,
			ConfigPath: configPath,
		})

		_, err := executeCmd(rootCmd, "config", "providers", "add", "--name", "gitlab", "--path", "/tmp/gitlab")

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

		_, err = executeCmd(rootCmd, "config", "providers", "remove", "--name", "gitlab")

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

		output, err := executeCmd(rootCmd, "config", "providers", "list")

		assert.NoError(err)
		assert.Contains(output, "gitlab")
		assert.Contains(output, "/tmp/gitlab")
	})
})
