package cmd

import (
	"errors"
	"os"
	"strings"
	"unicode"

	"github.com/charmbracelet/huh"
	"github.com/louiss0/cobra-cli-template/custom_errors"
	"github.com/louiss0/cobra-cli-template/internal/cmdutil"
	"github.com/louiss0/cobra-cli-template/internal/config"
	"github.com/louiss0/cobra-cli-template/internal/packagepath"
	"github.com/louiss0/cobra-cli-template/internal/project"
	"github.com/louiss0/cobra-cli-template/internal/prompt"
	"github.com/louiss0/cobra-cli-template/internal/runner"
	"github.com/spf13/cobra"
)

func NewInitCmd(commandRunner runner.Runner, promptRunner prompt.Runner, configPath *string) *cobra.Command {
	var siteFlag string
	var userFlag string
	var allowFull bool

	cmd := &cobra.Command{
		Use:   "init [package]",
		Short: "Initialize a Go module with a short package name",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			promptValues := initPrompt{}
			moduleInput := ""

			if len(args) == 0 {
				cmdutil.LogInfoIfProduction("init: starting interactive prompt")
				inputs, err := promptInitInputs(cmd, promptRunner)
				if err != nil {
					if errors.Is(err, huh.ErrUserAborted) {
						return nil
					}
					return err
				}
				promptValues = inputs
				moduleInput = inputs.ModuleName
			} else {
				moduleInput = args[0]
			}

			if strings.TrimSpace(moduleInput) == "" {
				return custom_errors.CreateInvalidInputErrorWithMessage("module name is required")
			}

			cmdutil.LogInfoIfProduction("init: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			configChanged := false
			if promptValues.UserName != "" {
				values.User = promptValues.UserName
				configChanged = true
			}
			if promptValues.ProviderSite != "" {
				values.Site = promptValues.ProviderSite
				configChanged = true
			}
			if promptValues.ShouldPersistTestChoice() {
				values.Scaffold.WriteTests = promptValues.TestDrivenChoice == testChoiceYes
				configChanged = true
			}
			if configChanged {
				if err := config.Save(*configPath, values); err != nil {
					return err
				}
			}

			site := config.ResolveSite(siteFlag, values)
			user, err := config.ResolveUser(userFlag, values, site)
			if err != nil {
				if errors.Is(err, config.ErrMissingUser) {
					return custom_errors.CreateInvalidInputErrorWithMessage("missing user; run go-toolkit config set-user <user>")
				}
				return err
			}

			allowCustomSite := allowFull || (siteFlag == "" && values.Site != "") || (!config.IsKnownSite(site) && site != "")
			if err := cmdutil.ValidateSite(site, allowCustomSite); err != nil {
				return err
			}

			cmdutil.LogInfoIfProduction("init: resolving module path for %s", site)
			modulePath, err := packagepath.ResolveModulePath(moduleInput, site, user)
			if err != nil {
				return err
			}

			cmdutil.LogInfoIfProduction("init: running go mod init")
			if err := commandRunner.Run(cmd, "go", "mod", "init", modulePath); err != nil {
				return err
			}

			if len(promptValues.Packages) > 0 {
				cmdutil.LogInfoIfProduction("init: installing packages")
				args := append([]string{"get"}, promptValues.Packages...)
				if err := commandRunner.Run(cmd, "go", args...); err != nil {
					return err
				}
			}

			writeMain := promptValues.ShouldWriteMain()
			shouldInitGit := promptValues.ShouldInitGit()

			cmdutil.LogInfoIfProduction("init: creating project layout")
			if err := project.EnsureLayout(".", project.Options{
				WriteMain:     writeMain,
				WriteInternal: true,
			}); err != nil {
				return err
			}

			if shouldInitGit {
				if _, err := os.Stat(".git"); err == nil {
					cmdutil.LogInfoIfProduction("init: git already initialized")
				} else if !errors.Is(err, os.ErrNotExist) {
					return err
				} else {
					cmdutil.LogInfoIfProduction("init: running git init")
					if err := commandRunner.Run(cmd, "git", "init"); err != nil {
						return err
					}
				}
			} else {
				cmdutil.LogInfoIfProduction("init: git init skipped")
			}

			return writeInitSummary(cmd, modulePath, site, user, promptValues)
		},
	}

	cmd.Flags().StringVar(&userFlag, "user", "", "override the configured user")
	cmd.Flags().StringVar(&siteFlag, "site", "", "override the configured site")
	cmd.Flags().BoolVar(&allowFull, "full", false, "allow a custom module site")
	cmdutil.RegisterSiteCompletion(cmd, "site")

	return cmd
}

const (
	projectTypeApp           = "app"
	projectTypeLibrary       = "library"
	projectTypeSkip          = "skip"
	projectTypeSkipRemaining = "skip-remaining"
	providerCustom           = "custom"
	providerSkip             = "skip"
	providerSkipRemaining    = "skip-remaining"
	testChoiceYes            = "yes"
	testChoiceNo             = "no"
	testChoiceSkip           = "skip"
	testChoiceSkipRemaining  = "skip-remaining"
	gitChoiceYes             = "yes"
	gitChoiceNo              = "no"
	gitChoiceSkip            = "skip"
	gitChoiceSkipRemaining   = "skip-remaining"
)

type initPrompt struct {
	ModuleName       string
	UserName         string
	ProviderSite     string
	ProjectType      string
	TestDrivenChoice string
	GitChoice        string
	Packages         []string
	Used             bool
}

func (p initPrompt) ShouldPersistTestChoice() bool {
	if !p.Used {
		return false
	}

	return p.TestDrivenChoice == testChoiceYes || p.TestDrivenChoice == testChoiceNo
}

func (p initPrompt) ShouldWriteMain() bool {
	if !p.Used {
		return true
	}

	if p.ProjectType == projectTypeLibrary {
		return false
	}

	return true
}

func (p initPrompt) ShouldInitGit() bool {
	if !p.Used {
		return true
	}

	if p.GitChoice == gitChoiceNo {
		return false
	}

	return true
}

type initSummary struct {
	ModulePath  string   `json:"module_path"`
	Site        string   `json:"site"`
	User        string   `json:"user"`
	ProjectType string   `json:"project_type"`
	TestDriven  string   `json:"test_driven"`
	GitInit     bool     `json:"git_init"`
	Packages    []string `json:"packages"`
}

func promptInitInputs(cmd *cobra.Command, runner prompt.Runner) (initPrompt, error) {
	moduleName, err := runner.Input(cmd, prompt.Input{
		Title:       "Module name",
		Placeholder: "go-toolkit",
		Validate: func(value string) error {
			if strings.TrimSpace(value) == "" {
				return errors.New("module name is required")
			}
			return nil
		},
	})
	if err != nil {
		return initPrompt{}, err
	}

	promptValues := initPrompt{
		ModuleName: moduleName,
		Used:       true,
	}

	userName, err := runner.Input(cmd, prompt.Input{
		Title:       "Username",
		Description: "Optional; leave blank to keep current config.",
		Placeholder: "lou",
	})
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return promptValues, nil
		}
		return initPrompt{}, err
	}

	promptValues.UserName = strings.TrimSpace(userName)

	providerChoice, err := runner.Select(cmd, prompt.Select{
		Title:   "Provider",
		Options: buildProviderOptions(),
	})
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return promptValues, nil
		}
		return initPrompt{}, err
	}

	if providerChoice == providerSkipRemaining {
		return promptValues, nil
	}

	if providerChoice == providerCustom {
		customSite, err := runner.Input(cmd, prompt.Input{
			Title:       "Custom provider",
			Placeholder: "github.com",
			Validate: func(value string) error {
				if strings.TrimSpace(value) == "" {
					return errors.New("provider is required")
				}
				return cmdutil.ValidateSite(value, true)
			},
		})
		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return promptValues, nil
			}
			return initPrompt{}, err
		}
		promptValues.ProviderSite = strings.TrimSpace(customSite)
	} else if providerChoice != providerSkip {
		promptValues.ProviderSite = providerChoice
	}

	projectType, err := runner.Select(cmd, prompt.Select{
		Title: "Project type",
		Options: []prompt.Option{
			{Label: "App", Value: projectTypeApp},
			{Label: "Library", Value: projectTypeLibrary},
			{Label: "Skip", Value: projectTypeSkip},
			{Label: "Skip remaining", Value: projectTypeSkipRemaining},
		},
	})
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return promptValues, nil
		}
		return initPrompt{}, err
	}

	if projectType == projectTypeSkipRemaining {
		return promptValues, nil
	}

	promptValues.ProjectType = projectType

	testChoice, err := runner.Select(cmd, prompt.Select{
		Title: "Test driven",
		Options: []prompt.Option{
			{Label: "Yes", Value: testChoiceYes},
			{Label: "No", Value: testChoiceNo},
			{Label: "Skip", Value: testChoiceSkip},
			{Label: "Skip remaining", Value: testChoiceSkipRemaining},
		},
	})
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return promptValues, nil
		}
		return initPrompt{}, err
	}

	if testChoice == testChoiceSkipRemaining {
		promptValues.TestDrivenChoice = testChoice
		return promptValues, nil
	}

	promptValues.TestDrivenChoice = testChoice

	gitChoice, err := runner.Select(cmd, prompt.Select{
		Title: "Use git init",
		Options: []prompt.Option{
			{Label: "Yes", Value: gitChoiceYes},
			{Label: "No", Value: gitChoiceNo},
			{Label: "Skip", Value: gitChoiceSkip},
			{Label: "Skip remaining", Value: gitChoiceSkipRemaining},
		},
	})
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return promptValues, nil
		}
		return initPrompt{}, err
	}

	if gitChoice == gitChoiceSkipRemaining {
		promptValues.GitChoice = gitChoice
		return promptValues, nil
	}

	promptValues.GitChoice = gitChoice

	packageInput, err := runner.Input(cmd, prompt.Input{
		Title:       "Packages to install",
		Description: "Space or comma separated module paths; leave blank to skip.",
		Placeholder: "github.com/spf13/cobra, github.com/spf13/viper",
	})
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return promptValues, nil
		}
		return initPrompt{}, err
	}

	promptValues.Packages = parsePackageList(packageInput)
	return promptValues, nil
}

func parsePackageList(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})

	if len(parts) == 0 {
		return nil
	}

	packages := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		packages = append(packages, trimmed)
	}

	return packages
}

func writeInitSummary(cmd *cobra.Command, modulePath string, site string, user string, prompt initPrompt) error {
	projectType := prompt.ProjectType
	if projectType == "" {
		projectType = projectTypeApp
	}

	testDriven := prompt.TestDrivenChoice
	if testDriven == "" {
		testDriven = testChoiceSkip
	}

	packages := prompt.Packages
	if packages == nil {
		packages = []string{}
	}

	summary := initSummary{
		ModulePath:  modulePath,
		Site:        site,
		User:        user,
		ProjectType: projectType,
		TestDriven:  testDriven,
		GitInit:     prompt.ShouldInitGit(),
		Packages:    packages,
	}

	return cmdutil.WritePrettyJSON(cmd.OutOrStdout(), summary)
}
