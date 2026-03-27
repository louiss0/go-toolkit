package cmd

import (
	"errors"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/louiss0/go-toolkit/custom_errors"
	"github.com/louiss0/go-toolkit/custom_flags"
	"github.com/louiss0/go-toolkit/internal/cmdutil"
	"github.com/louiss0/go-toolkit/internal/modindex/config"
	"github.com/louiss0/go-toolkit/internal/packagepath"
	"github.com/louiss0/go-toolkit/internal/project"
	"github.com/louiss0/go-toolkit/internal/prompt"
	"github.com/louiss0/go-toolkit/internal/runner"
	"github.com/louiss0/go-toolkit/validation"
	"github.com/spf13/cobra"
)

func NewInitCmd(commandRunner runner.Runner, promptRunner prompt.Runner, configPath *string) *cobra.Command {
	siteFlag := custom_flags.NewEmptyStringFlag("site")
	userFlag := custom_flags.NewEmptyStringFlag("user")
	templateFlag := custom_flags.NewUnionFlag(project.TemplateValues(), "template")
	var allowFull bool
	var packageFlags []string
	var presetFlags []string

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

			if _, err := validation.RequiredString(moduleInput, "module name"); err != nil {
				return err
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

			site := config.ResolveSite(siteFlag.String(), values)
			user, err := config.ResolveUser(userFlag.String(), values, site)
			if err != nil {
				if errors.Is(err, config.ErrMissingUser) {
					return custom_errors.CreateInvalidInputErrorWithMessage("missing user; run go-toolkit config set-user <user>")
				}
				return err
			}

			allowCustomSite := allowFull || (siteFlag.String() == "" && values.Site != "") || (!config.IsKnownSite(site) && site != "")
			if err := cmdutil.ValidateSite(site, allowCustomSite); err != nil {
				return err
			}

			installPackages, err := resolveInstallPackages(values, packageFlags, presetFlags, promptValues.Packages)
			if err != nil {
				return err
			}
			installPackages, err = assurePackageProviders(cmd, promptRunner, values, site, installPackages)
			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					return nil
				}
				return err
			}
			installPackages, err = resolveModulePaths(installPackages, site, user)
			if err != nil {
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

			if len(installPackages) > 0 {
				cmdutil.LogInfoIfProduction("init: installing packages")
				args := append([]string{"get"}, installPackages...)
				if err := commandRunner.Run(cmd, "go", args...); err != nil {
					return err
				}
			}

			template := resolveInitTemplate(templateFlag.String(), promptValues)
			shouldInitGit := promptValues.ShouldInitGit()

			cmdutil.LogInfoIfProduction("init: creating project layout from %s template", template)
			if err := project.EnsureLayout(".", project.Options{
				Template: template,
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

			promptValues.Packages = installPackages

			return writeInitSummary(cmd, modulePath, site, user, promptValues)
		},
	}

	cmd.Flags().Var(&userFlag, "user", "override the configured user")
	cmd.Flags().Var(&siteFlag, "site", "override the configured site")
	cmd.Flags().BoolVar(&allowFull, "full", false, "allow a custom module site")
	cmd.Flags().Var(&templateFlag, "template", "project template to apply")
	cmd.Flags().StringSliceVar(&packageFlags, "package", nil, "module paths to install after init")
	cmd.Flags().StringSliceVar(&presetFlags, "preset", nil, "package preset names to install after init")
	cmdutil.RegisterSiteCompletion(cmd, "site")
	_ = cmd.RegisterFlagCompletionFunc("template", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return project.TemplateValues(), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

const (
	templateTypeAPI           = project.TemplateAPI
	templateTypeCLI           = project.TemplateCLI
	templateTypeLib           = project.TemplateLib
	templateTypeSkip          = "skip"
	templateTypeSkipRemaining = "skip-remaining"
	providerCustom            = "custom"
	providerSkip              = "skip"
	providerSkipRemaining     = "skip-remaining"
	testChoiceYes             = "yes"
	testChoiceNo              = "no"
	testChoiceSkip            = "skip"
	testChoiceSkipRemaining   = "skip-remaining"
	gitChoiceYes              = "yes"
	gitChoiceNo               = "no"
	gitChoiceSkip             = "skip"
	gitChoiceSkipRemaining    = "skip-remaining"
)

type initPrompt struct {
	ModuleName       string
	UserName         string
	ProviderSite     string
	TemplateType     string
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
			_, err := validation.RequiredString(value, "module name")
			return err
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
				trimmed, err := validation.RequiredString(value, "provider")
				if err != nil {
					return err
				}
				return cmdutil.ValidateSite(trimmed, true)
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
		Title: "Template",
		Options: []prompt.Option{
			{Label: "API", Value: templateTypeAPI},
			{Label: "CLI", Value: templateTypeCLI},
			{Label: "Lib", Value: templateTypeLib},
			{Label: "Skip", Value: templateTypeSkip},
			{Label: "Skip remaining", Value: templateTypeSkipRemaining},
		},
	})
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return promptValues, nil
		}
		return initPrompt{}, err
	}

	if projectType == templateTypeSkipRemaining {
		return promptValues, nil
	}

	if projectType != templateTypeSkip {
		promptValues.TemplateType = projectType
	}

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
		Description: "Optional; use space-separated username/package or username/package/vN entries, or leave blank to skip.",
		Placeholder: "samber/lo onsi/ginkgo/v2",
		Validate: func(value string) error {
			_, err := validation.ParseShortPackageList(value, "packages to install")
			return err
		},
	})
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return promptValues, nil
		}
		return initPrompt{}, err
	}

	promptValues.Packages, err = validation.ParseShortPackageList(packageInput, "packages to install")
	if err != nil {
		return initPrompt{}, err
	}
	return promptValues, nil
}

func writeInitSummary(cmd *cobra.Command, modulePath string, site string, user string, prompt initPrompt) error {
	projectType := prompt.TemplateType
	if projectType == "" {
		projectType = templateTypeAPI
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

func resolveInitTemplate(flagValue string, promptValues initPrompt) string {
	if flagValue != "" {
		return flagValue
	}

	if promptValues.TemplateType != "" {
		return promptValues.TemplateType
	}

	return templateTypeAPI
}
