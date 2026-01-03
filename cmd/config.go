package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/louiss0/cobra-cli-template/custom_errors"
	"github.com/louiss0/cobra-cli-template/internal/cmdutil"
	"github.com/louiss0/cobra-cli-template/internal/config"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func NewConfigCmd(configPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage go-toolkit configuration",
	}

	cmd.AddCommand(newConfigInitCmd(configPath))
	cmd.AddCommand(newConfigShowCmd(configPath))
	cmd.AddCommand(newConfigSetUserCmd(configPath))
	cmd.AddCommand(newConfigSetSiteCmd(configPath))
	cmd.AddCommand(newConfigProvidersCmd(configPath))

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

func newConfigInitCmd(configPath *string) *cobra.Command {
	var userFlag string
	var siteFlag string
	var allowFull bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogInfoIfProduction("config init: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			if userFlag != "" {
				values.User = userFlag
			}

			if siteFlag != "" {
				values.Site = siteFlag
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

			fmt.Fprintln(cmd.OutOrStdout(), "config initialized")
			return nil
		},
	}

	cmd.Flags().StringVar(&userFlag, "user", "", "set the default user")
	cmd.Flags().StringVar(&siteFlag, "site", "", "set the default site")
	cmd.Flags().BoolVar(&allowFull, "full", false, "allow a custom module site")
	cmdutil.RegisterSiteCompletion(cmd, "site")

	return cmd
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

			site := config.ResolveSite("", values)
			user, err := config.ResolveUser("", values, site)
			if err != nil && !errors.Is(err, config.ErrMissingUser) {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "path: %s\n", *configPath)
			fmt.Fprintf(cmd.OutOrStdout(), "site: %s\n", site)
			fmt.Fprintf(cmd.OutOrStdout(), "user: %s\n", user)
			if len(values.Providers) > 0 {
				for _, provider := range values.Providers {
					fmt.Fprintf(cmd.OutOrStdout(), "provider: %s %s\n", provider.Name, provider.Path)
				}
			}
			return nil
		},
	}
}

func newConfigProvidersCmd(configPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "providers",
		Short: "Manage provider config mappings",
	}

	cmd.AddCommand(newConfigProvidersAddCmd(configPath))
	cmd.AddCommand(newConfigProvidersListCmd(configPath))
	cmd.AddCommand(newConfigProvidersRemoveCmd(configPath))

	return cmd
}

func newConfigProvidersAddCmd(configPath *string) *cobra.Command {
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

func newConfigProvidersRemoveCmd(configPath *string) *cobra.Command {
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

func newConfigProvidersListCmd(configPath *string) *cobra.Command {
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
