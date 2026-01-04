package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/louiss0/cobra-cli-template/custom_errors"
	"github.com/louiss0/cobra-cli-template/internal/cmdutil"
	"github.com/louiss0/cobra-cli-template/internal/config"
	"github.com/louiss0/cobra-cli-template/internal/prompt"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func NewConfigCmd(configPath *string, promptRunner prompt.Runner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage go-toolkit configuration",
	}

	cmd.AddCommand(newConfigInitCmd(configPath, promptRunner))
	cmd.AddCommand(newConfigShowCmd(configPath))
	cmd.AddCommand(newConfigSetUserCmd(configPath))
	cmd.AddCommand(newConfigSetSiteCmd(configPath))
	cmd.AddCommand(newConfigSetScaffoldTestsCmd(configPath))
	cmd.AddCommand(newConfigProviderCmd(configPath))

	return cmd
}

func newConfigSetUserCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "set-user <user>",
		Short: "Register the default user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogInfoIfProduction("config set-user: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			values.User = args[0]
			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "user saved")
			return nil
		},
	}
}

func newConfigSetSiteCmd(configPath *string) *cobra.Command {
	var allowFull bool

	cmd := &cobra.Command{
		Use:   "set-site <site>",
		Short: "Register the default module site",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogInfoIfProduction("config set-site: validating site")
			if err := cmdutil.ValidateSite(args[0], allowFull); err != nil {
				return err
			}

			cmdutil.LogInfoIfProduction("config set-site: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			values.Site = args[0]
			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "site saved")
			return nil
		},
	}

	cmd.Flags().BoolVar(&allowFull, "full", false, "allow a custom module site")
	cmdutil.RegisterSiteCompletion(cmd, "site")

	return cmd
}

func newConfigInitCmd(configPath *string, promptRunner prompt.Runner) *cobra.Command {
	var userFlag string
	var siteFlag string
	var allowFull bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			promptValues := configInitPrompt{}
			if userFlag == "" && siteFlag == "" {
				inputs, err := promptConfigInitInputs(cmd, promptRunner)
				if err != nil {
					if errors.Is(err, huh.ErrUserAborted) {
						return nil
					}
					return err
				}
				promptValues = inputs
			}

			cmdutil.LogInfoIfProduction("config init: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			if userFlag != "" {
				values.User = userFlag
			} else if promptValues.UserName != "" {
				values.User = promptValues.UserName
			}

			if siteFlag != "" {
				values.Site = siteFlag
			} else if promptValues.ProviderSite != "" {
				values.Site = promptValues.ProviderSite
			}

			if values.Site == "" {
				values.Site = config.DefaultSite
			}

			cmdutil.LogInfoIfProduction("config init: validating site")
			if err := cmdutil.ValidateSite(values.Site, allowFull); err != nil {
				return err
			}

			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			return writeConfigSummary(cmd, *configPath, values)
		},
	}

	cmd.Flags().StringVar(&userFlag, "user", "", "set the default user")
	cmd.Flags().StringVar(&siteFlag, "site", "", "set the default site")
	cmd.Flags().BoolVar(&allowFull, "full", false, "allow a custom module site")
	cmdutil.RegisterSiteCompletion(cmd, "site")

	return cmd
}

type configInitPrompt struct {
	UserName     string
	ProviderSite string
}

type configSummary struct {
	Path      string                  `json:"path"`
	Site      string                  `json:"site"`
	User      string                  `json:"user"`
	Scaffold  config.ScaffoldConfig   `json:"scaffold"`
	Providers []config.ProviderConfig `json:"providers"`
}

func promptConfigInitInputs(cmd *cobra.Command, runner prompt.Runner) (configInitPrompt, error) {
	userName, err := runner.Input(cmd, prompt.Input{
		Title:       "Username",
		Placeholder: "lou",
		Validate: func(value string) error {
			if strings.TrimSpace(value) == "" {
				return errors.New("username is required")
			}
			return nil
		},
	})
	if err != nil {
		return configInitPrompt{}, err
	}

	promptValues := configInitPrompt{
		UserName: strings.TrimSpace(userName),
	}

	providerChoice, err := runner.Select(cmd, prompt.Select{
		Title:   "Provider",
		Options: buildProviderOptions(),
	})
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return promptValues, nil
		}
		return configInitPrompt{}, err
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
			return configInitPrompt{}, err
		}
		promptValues.ProviderSite = strings.TrimSpace(customSite)
	} else if providerChoice != providerSkip {
		promptValues.ProviderSite = providerChoice
	}

	return promptValues, nil
}

func buildProviderOptions() []prompt.Option {
	knownSites := config.KnownSites()
	options := make([]prompt.Option, 0, len(knownSites)+3)
	for _, site := range knownSites {
		options = append(options, prompt.Option{Label: site, Value: site})
	}
	options = append(options,
		prompt.Option{Label: "Custom", Value: providerCustom},
		prompt.Option{Label: "Skip", Value: providerSkip},
		prompt.Option{Label: "Skip remaining", Value: providerSkipRemaining},
	)
	return options
}

func buildConfigSummary(configPath string, values config.Values) (configSummary, error) {
	site := config.ResolveSite("", values)
	user, err := config.ResolveUser("", values, site)
	if err != nil && !errors.Is(err, config.ErrMissingUser) {
		return configSummary{}, err
	}

	providers := values.Providers
	if providers == nil {
		providers = []config.ProviderConfig{}
	}

	return configSummary{
		Path:      configPath,
		Site:      site,
		User:      user,
		Scaffold:  values.Scaffold,
		Providers: providers,
	}, nil
}

func writeConfigSummary(cmd *cobra.Command, configPath string, values config.Values) error {
	summary, err := buildConfigSummary(configPath, values)
	if err != nil {
		return err
	}
	return cmdutil.WritePrettyJSON(cmd.OutOrStdout(), summary)
}

func newConfigShowCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show the current config values",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogInfoIfProduction("config show: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			summary, err := buildConfigSummary(*configPath, values)
			if err != nil {
				return err
			}
			return cmdutil.WritePrettyJSON(cmd.OutOrStdout(), summary)
		},
	}
}

func newConfigSetScaffoldTestsCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "set-scaffold-tests <enabled>",
		Short: "Enable or disable scaffold test generation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			enabled, err := strconv.ParseBool(args[0])
			if err != nil {
				return custom_errors.CreateInvalidInputErrorWithMessage("enabled must be true or false")
			}

			cmdutil.LogInfoIfProduction("config set-scaffold-tests: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			values.Scaffold.WriteTests = enabled
			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "scaffold tests updated")
			return nil
		},
	}
}

func newConfigProviderCmd(configPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage provider config mappings",
	}

	cmd.AddCommand(newConfigProviderAddCmd(configPath))
	cmd.AddCommand(newConfigProviderListCmd(configPath))
	cmd.AddCommand(newConfigProviderRemoveCmd(configPath))

	return cmd
}

func newConfigProviderAddCmd(configPath *string) *cobra.Command {
	var nameFlag string
	var pathFlag string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a provider config mapping",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(nameFlag) == "" {
				return custom_errors.CreateInvalidInputErrorWithMessage("provider name is required")
			}
			if strings.TrimSpace(pathFlag) == "" {
				return custom_errors.CreateInvalidInputErrorWithMessage("provider path is required")
			}

			cmdutil.LogInfoIfProduction("config providers add: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			entry := config.ProviderConfig{
				Name: nameFlag,
				Path: pathFlag,
			}

			values.Providers = append(values.Providers, entry)
			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "provider added")
			return nil
		},
	}

	cmd.Flags().StringVar(&nameFlag, "name", "", "provider name")
	cmd.Flags().StringVar(&pathFlag, "path", "", "path to provider config")

	return cmd
}

func newConfigProviderRemoveCmd(configPath *string) *cobra.Command {
	var nameFlag string

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a provider config mapping",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(nameFlag) == "" {
				return custom_errors.CreateInvalidInputErrorWithMessage("provider name is required")
			}

			cmdutil.LogInfoIfProduction("config providers remove: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			filtered := lo.Filter(values.Providers, func(item config.ProviderConfig, _ int) bool {
				return item.Name != nameFlag
			})

			if len(filtered) == len(values.Providers) {
				return custom_errors.CreateInvalidInputErrorWithMessage("provider name not found")
			}

			values.Providers = filtered
			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "provider removed")
			return nil
		},
	}

	cmd.Flags().StringVar(&nameFlag, "name", "", "provider name")

	return cmd
}

func newConfigProviderListCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List provider config mappings",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogInfoIfProduction("config providers list: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			for _, provider := range values.Providers {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", provider.Name, provider.Path)
			}

			return nil
		},
	}
}
