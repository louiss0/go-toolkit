package main_test

import (
	"bytes"
	"context"
	"encoding/json"
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

func (M *runnerMock) Run(Name string, Args []string, Stdout, Stderr io.Writer) error {
	Call := M.Called(Name, Args, Stdout, Stderr)
	return Call.Error(0)
}

type indexFetcherMock struct {
	mock.Mock
}

func (M *indexFetcherMock) Fetch(Ctx context.Context, Request modindex.Request) ([]modindex.Entry, error) {
	Call := M.Called(Ctx, Request)
	Entries, _ := Call.Get(0).([]modindex.Entry)
	return Entries, Call.Error(1)
}

func ExecuteCmd(Cmd *cobra.Command, Args ...string) (string, error) {
	Buf := new(bytes.Buffer)
	ErrBuff := new(bytes.Buffer)

	Cmd.SetOut(Buf)
	Cmd.SetErr(ErrBuff)
	Cmd.SetArgs(Args)

	Err := Cmd.Execute()
	if ErrBuff.Len() > 0 {
		return "", fmt.Errorf("command failed: %s", ErrBuff.String())
	}

	return Buf.String(), Err
}

var Test = Describe("test command", func() {
	Assert := assert.New(GinkgoT())

	It("runs go test for all packages by default", func() {
		Runner := &runnerMock{}
		Runner.On("Run", "go", []string{"test", "./..."}, mock.Anything, mock.Anything).Return(nil).Once()

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: Runner,
		})

		_, Err := ExecuteCmd(RootCmd, "test")

		Assert.NoError(Err)
		Runner.AssertExpectations(GinkgoT())
	})
})

var Remove = Describe("remove command", func() {
	Assert := assert.New(GinkgoT())

	It("removes a fully qualified module", func() {
		Runner := &runnerMock{}
		Runner.On("Run", "go", []string{"get", "github.com/acme/tool@none"}, mock.Anything, mock.Anything).Return(nil).Once()

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: Runner,
		})

		_, Err := ExecuteCmd(RootCmd, "remove", "github.com/acme/tool")

		Assert.NoError(Err)
		Runner.AssertExpectations(GinkgoT())
	})

	It("removes multiple modules in one command", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		ConfigPath := filepath.Join(TempDir, "config.toml")

		Err := os.WriteFile(ConfigPath, []byte("user = \"lou\"\nsite = \"github.com\"\n"), 0o644)
		Assert.NoError(Err)

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     Runner,
			ConfigPath: ConfigPath,
		})

		Runner.On(
			"Run",
			"go",
			[]string{"get", "github.com/lou/tool@none", "github.com/acme/other@none"},
			mock.Anything,
			mock.Anything,
		).Return(nil).Once()

		_, Err = ExecuteCmd(RootCmd, "remove", "tool", "github.com/acme/other")

		Assert.NoError(Err)
		Runner.AssertExpectations(GinkgoT())
	})

	It("removes a module path with a version suffix", func() {
		Runner := &runnerMock{}
		Runner.On("Run", "go", []string{"get", "github.com/onsi/ginkgo/v2@none"}, mock.Anything, mock.Anything).Return(nil).Once()

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: Runner,
		})

		_, Err := ExecuteCmd(RootCmd, "remove", "github.com/onsi/ginkgo/v2")

		Assert.NoError(Err)
		Runner.AssertExpectations(GinkgoT())
	})

	It("rejects @none for remove input", func() {
		Runner := &runnerMock{}

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: Runner,
		})

		_, Err := ExecuteCmd(RootCmd, "remove", "github.com/onsi/ginkgo@none")

		Assert.Error(Err)
		Assert.Contains(Err.Error(), "added automatically")
		Runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	It("prints the remove command on dry run", func() {
		Runner := &runnerMock{}

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: Runner,
		})

		Output, Err := ExecuteCmd(RootCmd, "remove", "github.com/onsi/ginkgo/v2", "--dry-run")

		Assert.NoError(Err)
		Runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		Assert.Contains(Output, "go get github.com/onsi/ginkgo/v2@none")
	})
})

var Init = Describe("init command", func() {
	Assert := assert.New(GinkgoT())

	It("inits a module using the registered user", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		ConfigPath := filepath.Join(TempDir, "config.toml")

		Err := os.WriteFile(ConfigPath, []byte("user = \"lou\"\nsite = \"github.com\"\n"), 0o644)
		Assert.NoError(Err)

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     Runner,
			ConfigPath: ConfigPath,
		})

		Runner.On("Run", "go", []string{"mod", "init", "github.com/lou/toolkit"}, mock.Anything, mock.Anything).Return(nil).Once()

		_, Err = ExecuteCmd(RootCmd, "init", "toolkit")

		Assert.NoError(Err)
		Runner.AssertExpectations(GinkgoT())
	})
})

var Search = Describe("search command", func() {
	Assert := assert.New(GinkgoT())

	It("searches the module index with a default site filter", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		ConfigPath := filepath.Join(TempDir, "config.toml")
		IndexFetcher := &indexFetcherMock{}
		Entries := []modindex.Entry{
			{Path: "github.com/acme/tool", Version: "v1.0.0"},
			{Path: "gitlab.com/acme/tool", Version: "v1.2.0"},
		}
		IndexFetcher.On("Fetch", mock.Anything, modindex.Request{Limit: 200}).Return(Entries, nil).Once()

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       Runner,
			IndexFetcher: IndexFetcher,
			ConfigPath:   ConfigPath,
		})

		Output, Err := ExecuteCmd(RootCmd, "search", "tool")

		Assert.NoError(Err)
		Assert.Contains(Output, "github.com/acme/tool")
		Assert.NotContains(Output, "gitlab.com/acme/tool")
		IndexFetcher.AssertExpectations(GinkgoT())
	})

	It("searches without a site filter when the query includes a domain", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		ConfigPath := filepath.Join(TempDir, "config.toml")
		IndexFetcher := &indexFetcherMock{}
		Entries := []modindex.Entry{
			{Path: "github.com/acme/tool", Version: "v1.0.0"},
			{Path: "gitlab.com/acme/tool", Version: "v1.2.0"},
		}
		IndexFetcher.On("Fetch", mock.Anything, modindex.Request{Limit: 200}).Return(Entries, nil).Once()

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       Runner,
			IndexFetcher: IndexFetcher,
			ConfigPath:   ConfigPath,
		})

		Output, Err := ExecuteCmd(RootCmd, "search", "gitlab.com")

		Assert.NoError(Err)
		Assert.Contains(Output, "gitlab.com/acme/tool")
		Assert.NotContains(Output, "github.com/acme/tool")
		IndexFetcher.AssertExpectations(GinkgoT())
	})

	It("prints details when requested", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		ConfigPath := filepath.Join(TempDir, "config.toml")
		IndexFetcher := &indexFetcherMock{}
		Entries := []modindex.Entry{
			{Path: "github.com/acme/tool", Version: "v1.2.3", Timestamp: "2024-01-01T00:00:00Z"},
		}
		IndexFetcher.On("Fetch", mock.Anything, modindex.Request{Limit: 200}).Return(Entries, nil).Once()

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       Runner,
			IndexFetcher: IndexFetcher,
			ConfigPath:   ConfigPath,
		})

		Output, Err := ExecuteCmd(RootCmd, "search", "tool", "--details")

		Assert.NoError(Err)
		Assert.Contains(Output, "github.com/acme/tool")
		Assert.Contains(Output, "v1.2.3")
		Assert.Contains(Output, "2024-01-01T00:00:00Z")
		IndexFetcher.AssertExpectations(GinkgoT())
	})

	It("rejects unknown sites without --full", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		ConfigPath := filepath.Join(TempDir, "config.toml")
		IndexFetcher := &indexFetcherMock{}

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       Runner,
			IndexFetcher: IndexFetcher,
			ConfigPath:   ConfigPath,
		})

		_, Err := ExecuteCmd(RootCmd, "search", "tool", "--site", "example.com")

		Assert.Error(Err)
		Assert.Contains(Err.Error(), "unsupported site")
		IndexFetcher.AssertNotCalled(GinkgoT(), "Fetch", mock.Anything, mock.Anything)
	})

	It("rejects sites without a dot", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		ConfigPath := filepath.Join(TempDir, "config.toml")
		IndexFetcher := &indexFetcherMock{}

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       Runner,
			IndexFetcher: IndexFetcher,
			ConfigPath:   ConfigPath,
		})

		_, Err := ExecuteCmd(RootCmd, "search", "tool", "--site", "githubcom")

		Assert.Error(Err)
		Assert.Contains(Err.Error(), "sitename.domain")
		IndexFetcher.AssertNotCalled(GinkgoT(), "Fetch", mock.Anything, mock.Anything)
	})

	It("outputs JSON when requested", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		ConfigPath := filepath.Join(TempDir, "config.toml")
		IndexFetcher := &indexFetcherMock{}
		Entries := []modindex.Entry{
			{Path: "github.com/acme/tool", Version: "v1.0.0", Timestamp: "2024-01-01T00:00:00Z"},
			{Path: "github.com/acme/tool", Version: "v1.1.0", Timestamp: "2024-01-02T00:00:00Z"},
		}
		IndexFetcher.On("Fetch", mock.Anything, modindex.Request{Limit: 200}).Return(Entries, nil).Once()

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       Runner,
			IndexFetcher: IndexFetcher,
			ConfigPath:   ConfigPath,
		})

		Output, Err := ExecuteCmd(RootCmd, "search", "tool", "--json")

		Assert.NoError(Err)
		var Payload []modindex.Entry
		Assert.NoError(json.Unmarshal([]byte(Output), &Payload))
		Assert.Len(Payload, 1)
		Assert.Equal("github.com/acme/tool", Payload[0].Path)
		Assert.Equal("v1.0.0", Payload[0].Version)
	})

	It("rejects since with since-hours", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		ConfigPath := filepath.Join(TempDir, "config.toml")
		IndexFetcher := &indexFetcherMock{}

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       Runner,
			IndexFetcher: IndexFetcher,
			ConfigPath:   ConfigPath,
		})

		_, Err := ExecuteCmd(RootCmd, "search", "tool", "--since", "2024-01-01T00:00:00Z", "--since-hours", "1")

		Assert.Error(Err)
		Assert.Contains(Err.Error(), "since cannot be combined")
		IndexFetcher.AssertNotCalled(GinkgoT(), "Fetch", mock.Anything, mock.Anything)
	})
})

var Scaffold = Describe("scaffold command", func() {
	Assert := assert.New(GinkgoT())

	It("scaffolds a folder with a README", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		Target := filepath.Join(TempDir, "demo")

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     Runner,
			ConfigPath: filepath.Join(TempDir, "config.toml"),
		})

		_, Err := ExecuteCmd(RootCmd, "scaffold", "demo", "--folder", Target, "--readme")

		Assert.NoError(Err)
		Runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		Content, Err := os.ReadFile(filepath.Join(Target, "README.md"))
		Assert.NoError(Err)
		Assert.Contains(string(Content), "# demo")
	})

	It("scaffolds and initializes a module", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		Target := filepath.Join(TempDir, "demo")
		ConfigPath := filepath.Join(TempDir, "config.toml")

		Err := os.WriteFile(ConfigPath, []byte("user = \"lou\"\nsite = \"github.com\"\n"), 0o644)
		Assert.NoError(Err)

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     Runner,
			ConfigPath: ConfigPath,
		})

		Runner.On("Run", "go", []string{"-C", Target, "mod", "init", "github.com/lou/demo"}, mock.Anything, mock.Anything).Return(nil).Once()

		_, Err = ExecuteCmd(RootCmd, "scaffold", "demo", "--folder", Target, "--module")

		Assert.NoError(Err)
		Runner.AssertExpectations(GinkgoT())
	})
})

var Add = Describe("add command", func() {
	Assert := assert.New(GinkgoT())

	It("adds multiple packages with short paths", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		ConfigPath := filepath.Join(TempDir, "config.toml")

		Err := os.WriteFile(ConfigPath, []byte("user = \"lou\"\nsite = \"github.com\"\n"), 0o644)
		Assert.NoError(Err)

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     Runner,
			ConfigPath: ConfigPath,
		})

		Runner.On(
			"Run",
			"go",
			[]string{"get", "github.com/samber/lo", "github.com/stretchr/testify", "github.com/onsi/ginkgo"},
			mock.Anything,
			mock.Anything,
		).Return(nil).Once()

		_, Err = ExecuteCmd(RootCmd, "add", "samber/lo", "stretchr/testify", "onsi/ginkgo")

		Assert.NoError(Err)
		Runner.AssertExpectations(GinkgoT())
	})

	It("adds a module path with a version suffix", func() {
		Runner := &runnerMock{}
		Runner.On("Run", "go", []string{"get", "github.com/onsi/ginkgo/v2"}, mock.Anything, mock.Anything).Return(nil).Once()

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: Runner,
		})

		_, Err := ExecuteCmd(RootCmd, "add", "github.com/onsi/ginkgo/v2")

		Assert.NoError(Err)
		Runner.AssertExpectations(GinkgoT())
	})

	It("adds a module with an @version suffix", func() {
		Runner := &runnerMock{}
		Runner.On("Run", "go", []string{"get", "github.com/onsi/ginkgo@v2.0.0"}, mock.Anything, mock.Anything).Return(nil).Once()

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: Runner,
		})

		_, Err := ExecuteCmd(RootCmd, "add", "github.com/onsi/ginkgo@v2.0.0")

		Assert.NoError(Err)
		Runner.AssertExpectations(GinkgoT())
	})

	It("rejects @none for add", func() {
		Runner := &runnerMock{}

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: Runner,
		})

		_, Err := ExecuteCmd(RootCmd, "add", "github.com/onsi/ginkgo@none")

		Assert.Error(Err)
		Assert.Contains(Err.Error(), "use remove")
		Runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	It("prints the add command on dry run", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		ConfigPath := filepath.Join(TempDir, "config.toml")

		Err := os.WriteFile(ConfigPath, []byte("user = \"lou\"\nsite = \"github.com\"\n"), 0o644)
		Assert.NoError(Err)

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     Runner,
			ConfigPath: ConfigPath,
		})

		Output, Err := ExecuteCmd(RootCmd, "add", "samber/lo", "--dry-run")

		Assert.NoError(Err)
		Runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		Assert.Contains(Output, "go get github.com/samber/lo")
	})
})

var Config = Describe("config command", func() {
	Assert := assert.New(GinkgoT())

	It("initializes config with defaults", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		ConfigPath := filepath.Join(TempDir, "config.toml")

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     Runner,
			ConfigPath: ConfigPath,
		})

		_, Err := ExecuteCmd(RootCmd, "config", "init", "--user", "lou")

		Assert.NoError(Err)

		Values, Err := config.Load(ConfigPath)
		Assert.NoError(Err)
		Assert.Equal("lou", Values.User)
		Assert.Equal("github.com", Values.Site)
	})

	It("shows config values", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		ConfigPath := filepath.Join(TempDir, "config.toml")

		Err := os.WriteFile(ConfigPath, []byte("user = \"lou\"\nsite = \"gitlab.com\"\n"), 0o644)
		Assert.NoError(Err)

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     Runner,
			ConfigPath: ConfigPath,
		})

		Output, Err := ExecuteCmd(RootCmd, "config", "show")

		Assert.NoError(Err)
		Assert.Contains(Output, "path:")
		Assert.Contains(Output, "site: gitlab.com")
		Assert.Contains(Output, "user: lou")
	})

	It("uses a repo-local gtk-config.toml when present", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		ConfigPath := filepath.Join(TempDir, "gtk-config.toml")

		CurrentDir, Err := os.Getwd()
		Assert.NoError(Err)

		Err = os.Chdir(TempDir)
		Assert.NoError(Err)
		defer func() {
			_ = os.Chdir(CurrentDir)
		}()

		Err = os.WriteFile(ConfigPath, []byte("user = \"lou\"\nsite = \"github.com\"\n"), 0o644)
		Assert.NoError(Err)

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner: Runner,
		})

		Output, Err := ExecuteCmd(RootCmd, "config", "show")

		Assert.NoError(Err)
		Assert.Contains(Output, "path: "+ConfigPath)
	})

	It("adds a provider mapping", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		ConfigPath := filepath.Join(TempDir, "config.toml")

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     Runner,
			ConfigPath: ConfigPath,
		})

		_, Err := ExecuteCmd(RootCmd, "config", "providers", "add", "--name", "gitlab", "--path", "/tmp/gitlab")

		Assert.NoError(Err)

		Values, Err := config.Load(ConfigPath)
		Assert.NoError(Err)
		Assert.Len(Values.Providers, 1)
		Assert.Equal("gitlab", Values.Providers[0].Name)
		Assert.Equal("/tmp/gitlab", Values.Providers[0].Path)
	})

	It("removes a provider mapping", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		ConfigPath := filepath.Join(TempDir, "config.toml")

		Err := os.WriteFile(ConfigPath, []byte("[[providers]]\nname = \"gitlab\"\npath = \"/tmp/gitlab\"\n"), 0o644)
		Assert.NoError(Err)

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     Runner,
			ConfigPath: ConfigPath,
		})

		_, Err = ExecuteCmd(RootCmd, "config", "providers", "remove", "--name", "gitlab")

		Assert.NoError(Err)

		Values, Err := config.Load(ConfigPath)
		Assert.NoError(Err)
		Assert.Empty(Values.Providers)
	})

	It("lists provider mappings", func() {
		Runner := &runnerMock{}
		TempDir := GinkgoT().TempDir()
		ConfigPath := filepath.Join(TempDir, "config.toml")

		Err := os.WriteFile(ConfigPath, []byte("[[providers]]\nname = \"gitlab\"\npath = \"/tmp/gitlab\"\n"), 0o644)
		Assert.NoError(Err)

		RootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:     Runner,
			ConfigPath: ConfigPath,
		})

		Output, Err := ExecuteCmd(RootCmd, "config", "providers", "list")

		Assert.NoError(Err)
		Assert.Contains(Output, "gitlab")
		Assert.Contains(Output, "/tmp/gitlab")
	})
})
