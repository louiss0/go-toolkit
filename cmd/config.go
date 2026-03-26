package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/carapace-sh/carapace-shlex"
	"github.com/charmbracelet/huh"
	"github.com/louiss0/go-toolkit/custom_errors"
	"github.com/louiss0/go-toolkit/custom_flags"
	"github.com/louiss0/go-toolkit/internal/cmdutil"
	"github.com/louiss0/go-toolkit/internal/modindex/config"
	"github.com/louiss0/go-toolkit/internal/prompt"
	"github.com/louiss0/go-toolkit/internal/runner"
	"github.com/louiss0/go-toolkit/validation"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func NewConfigCmd(commandRunner runner.Runner, configPath *string, promptRunner prompt.Runner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage go-toolkit configuration",
	}

	cmd.AddCommand(newConfigInitCmd(configPath, promptRunner))
	cmd.AddCommand(newConfigEditCmd(commandRunner, configPath))
	cmd.AddCommand(newConfigShowCmd(configPath))
	cmd.AddCommand(newConfigSetUserCmd(configPath))
	cmd.AddCommand(newConfigSetSiteCmd(configPath))
	cmd.AddCommand(newConfigSetAssureProvidersCmd(configPath))
	cmd.AddCommand(newConfigSetScaffoldTestsCmd(configPath))
	cmd.AddCommand(newConfigProviderCmd(configPath))
	cmd.AddCommand(newConfigPackagePresetCmd(configPath))
	cmd.AddCommand(newConfigGlobalPackageCmd(configPath))
	cmd.AddCommand(newConfigRemoveCmd(configPath))

	return cmd
}

func newConfigEditCmd(commandRunner runner.Runner, configPath *string) *cobra.Command {
	editorFlag := custom_flags.NewEmptyStringFlag("editor")

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Open the config file in your editor",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			targetPath := strings.TrimSpace(*configPath)
			if targetPath == "" {
				return custom_errors.CreateInvalidInputErrorWithMessage("config path is required")
			}
			if err := ensureConfigFileExists(targetPath); err != nil {
				return err
			}

			editorParts, err := resolveEditorCommand(editorFlag.String())
			if err != nil {
				return err
			}

			editorName := editorParts[0]
			editorArgs := append(editorParts[1:], targetPath)
			return commandRunner.Run(cmd, editorName, editorArgs...)
		},
	}

	cmd.Flags().Var(&editorFlag, "editor", "override the editor command")

	return cmd
}

func ensureConfigFileExists(path string) error {
	configDir := filepath.Dir(path)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return os.WriteFile(path, []byte{}, 0o644)
	} else if err != nil {
		return err
	}

	return nil
}

func resolveEditorCommand(override string) ([]string, error) {
	editorValue := strings.TrimSpace(override)
	if editorValue == "" {
		for _, candidate := range []string{
			os.Getenv("GO_TOOLKIT_EDITOR"),
			os.Getenv("GIT_EDITOR"),
			os.Getenv("VISUAL"),
			os.Getenv("EDITOR"),
		} {
			if strings.TrimSpace(candidate) != "" {
				editorValue = candidate
				break
			}
		}
	}
	if editorValue == "" {
		if runtime.GOOS == "windows" {
			return []string{"notepad"}, nil
		}
		return nil, custom_errors.CreateInvalidInputErrorWithMessage(
			"missing editor; set --editor, GO_TOOLKIT_EDITOR, GIT_EDITOR, VISUAL, or EDITOR",
		)
	}

	tokens, err := shlex.Split(editorValue)
	if err != nil {
		return nil, custom_errors.CreateInvalidInputErrorWithMessage("editor command is invalid")
	}

	parts := lo.Filter(tokens.Strings(), func(part string, _ int) bool {
		return strings.TrimSpace(part) != ""
	})
	if len(parts) == 0 {
		return nil, custom_errors.CreateInvalidInputErrorWithMessage("editor command is invalid")
	}

	return parts, nil
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

			user, err := validation.RequiredString(args[0], "user")
			if err != nil {
				return err
			}

			values.User = user
			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			return cmdutil.WriteLine(cmd.OutOrStdout(), "user saved")
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

			return cmdutil.WriteLine(cmd.OutOrStdout(), "site saved")
		},
	}

	cmd.Flags().BoolVar(&allowFull, "full", false, "allow a custom module site")
	cmdutil.RegisterSiteCompletion(cmd, "site")

	return cmd
}

func newConfigSetAssureProvidersCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "set-assure-providers <enabled>",
		Short: "Enable or disable provider assurance for short package paths",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			enabled, err := validation.ParseBool(args[0], "enabled")
			if err != nil {
				return err
			}

			cmdutil.LogInfoIfProduction("config set-assure-providers: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			values.AssureProviders = enabled
			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			return cmdutil.WriteLine(cmd.OutOrStdout(), "provider assurance updated")
		},
	}
}

func newConfigInitCmd(configPath *string, promptRunner prompt.Runner) *cobra.Command {
	userFlag := custom_flags.NewEmptyStringFlag("user")
	siteFlag := custom_flags.NewEmptyStringFlag("site")
	var allowFull bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			promptValues := configInitPrompt{}
			if userFlag.String() == "" && siteFlag.String() == "" {
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

			if userFlag.String() != "" {
				values.User = userFlag.String()
			} else if promptValues.UserName != "" {
				values.User = promptValues.UserName
			}

			if siteFlag.String() != "" {
				values.Site = siteFlag.String()
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

	cmd.Flags().Var(&userFlag, "user", "set the default user")
	cmd.Flags().Var(&siteFlag, "site", "set the default site")
	cmd.Flags().BoolVar(&allowFull, "full", false, "allow a custom module site")
	cmdutil.RegisterSiteCompletion(cmd, "site")

	return cmd
}

type configInitPrompt struct {
	UserName     string
	ProviderSite string
}

type configSummary struct {
	Path            string                  `json:"path"`
	Site            string                  `json:"site"`
	User            string                  `json:"user"`
	AssureProviders bool                    `json:"assure_providers"`
	Scaffold        config.ScaffoldConfig   `json:"scaffold"`
	Providers       []config.ProviderConfig `json:"providers"`
	PackagePresets  map[string][]string     `json:"package_presets"`
	GlobalPackages  []string                `json:"global_packages"`
}

func promptConfigInitInputs(cmd *cobra.Command, runner prompt.Runner) (configInitPrompt, error) {
	userName, err := runner.Input(cmd, prompt.Input{
		Title:       "Username",
		Placeholder: "lou",
		Validate: func(value string) error {
			_, err := validation.RequiredString(value, "username")
			return err
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
	options := lo.Map(knownSites, func(site string, _ int) prompt.Option {
		label := site
		if knownLabel, ok := config.KnownSiteLabel(site); ok {
			label = knownLabel
		}
		return prompt.Option{Label: label, Value: site}
	})

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
	packagePresets := values.PackagePresets
	if packagePresets == nil {
		packagePresets = map[string][]string{}
	}
	globalPackages := values.GlobalPackages
	if globalPackages == nil {
		globalPackages = []string{}
	}

	return configSummary{
		Path:            configPath,
		Site:            site,
		User:            user,
		AssureProviders: values.AssureProviders,
		Scaffold:        values.Scaffold,
		Providers:       providers,
		PackagePresets:  packagePresets,
		GlobalPackages:  globalPackages,
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
			enabled, err := validation.ParseBool(args[0], "enabled")
			if err != nil {
				return err
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

			return cmdutil.WriteLine(cmd.OutOrStdout(), "scaffold tests updated")
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

func newConfigPackagePresetCmd(configPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package-preset",
		Short: "Manage package install presets",
	}

	cmd.AddCommand(newConfigPackagePresetAddCmd(configPath))
	cmd.AddCommand(newConfigPackagePresetListCmd(configPath))
	cmd.AddCommand(newConfigPackagePresetRemoveCmd(configPath))

	return cmd
}

func newConfigProviderAddCmd(configPath *string) *cobra.Command {
	nameFlag := custom_flags.NewEmptyStringFlag("name")
	pathFlag := custom_flags.NewEmptyStringFlag("path")

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a provider config mapping",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := validation.RequiredString(nameFlag.String(), "provider name")
			if err != nil {
				return err
			}
			path, err := validation.RequiredString(pathFlag.String(), "provider path")
			if err != nil {
				return err
			}

			cmdutil.LogInfoIfProduction("config providers add: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			entry := config.ProviderConfig{
				Name: name,
				Path: path,
			}

			values.Providers = append(values.Providers, entry)
			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			return cmdutil.WriteLine(cmd.OutOrStdout(), "provider added")
		},
	}

	cmd.Flags().Var(&nameFlag, "name", "provider name")
	cmd.Flags().Var(&pathFlag, "path", "path to provider config")

	return cmd
}

func newConfigProviderRemoveCmd(configPath *string) *cobra.Command {
	nameFlag := custom_flags.NewEmptyStringFlag("name")

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a provider config mapping",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := validation.RequiredString(nameFlag.String(), "provider name")
			if err != nil {
				return err
			}

			cmdutil.LogInfoIfProduction("config providers remove: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			filtered := lo.Filter(values.Providers, func(item config.ProviderConfig, _ int) bool {
				return item.Name != name
			})

			if len(filtered) == len(values.Providers) {
				return custom_errors.CreateInvalidInputErrorWithMessage("provider name not found")
			}

			values.Providers = filtered
			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			return cmdutil.WriteLine(cmd.OutOrStdout(), "provider removed")
		},
	}

	cmd.Flags().Var(&nameFlag, "name", "provider name")

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

			rows := lo.Map(values.Providers, func(provider config.ProviderConfig, _ int) string {
				return fmt.Sprintf("%s\t%s", provider.Name, provider.Path)
			})
			if len(rows) > 0 {
				return cmdutil.WriteLine(cmd.OutOrStdout(), strings.Join(rows, "\n"))
			}

			return nil
		},
	}
}

func newConfigPackagePresetAddCmd(configPath *string) *cobra.Command {
	nameFlag := custom_flags.NewEmptyStringFlag("name")
	var packageFlags []string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a package install preset",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			name, err := validation.RequiredString(nameFlag.String(), "package preset name")
			if err != nil {
				return err
			}
			if len(packageFlags) == 0 {
				return custom_errors.CreateInvalidInputErrorWithMessage("at least one package is required")
			}
			trimmedPackages, err := validation.NonEmptyStrings(packageFlags, "package values")
			if err != nil {
				return err
			}
			nameFlag = custom_flags.NewEmptyStringFlag("name")
			if err := nameFlag.Set(name); err != nil {
				return err
			}
			packageFlags = trimmedPackages

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogInfoIfProduction("config package preset add: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			if values.PackagePresets == nil {
				values.PackagePresets = map[string][]string{}
			}
			values.PackagePresets[nameFlag.String()] = lo.Uniq(packageFlags)
			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			return cmdutil.WriteLine(cmd.OutOrStdout(), "package preset saved")
		},
	}

	cmd.Flags().Var(&nameFlag, "name", "preset name")
	cmd.Flags().StringSliceVar(&packageFlags, "package", nil, "packages included in the preset")

	return cmd
}

func newConfigPackagePresetRemoveCmd(configPath *string) *cobra.Command {
	nameFlag := custom_flags.NewEmptyStringFlag("name")

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a package install preset",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			name, err := validation.RequiredString(nameFlag.String(), "package preset name")
			if err != nil {
				return err
			}
			nameFlag = custom_flags.NewEmptyStringFlag("name")
			if err := nameFlag.Set(name); err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogInfoIfProduction("config package preset remove: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			if _, ok := values.PackagePresets[nameFlag.String()]; !ok {
				return custom_errors.CreateInvalidInputErrorWithMessage("package preset name not found")
			}

			delete(values.PackagePresets, nameFlag.String())
			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			return cmdutil.WriteLine(cmd.OutOrStdout(), "package preset removed")
		},
	}

	cmd.Flags().Var(&nameFlag, "name", "preset name")

	return cmd
}

func newConfigPackagePresetListCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List package install presets",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogInfoIfProduction("config package preset list: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			rows := lo.Map(config.KnownPackagePresetNames(values), func(name string, _ int) string {
				return fmt.Sprintf("%s\t%s", name, strings.Join(values.PackagePresets[name], ", "))
			})
			if len(rows) > 0 {
				return cmdutil.WriteLine(cmd.OutOrStdout(), strings.Join(rows, "\n"))
			}

			return nil
		},
	}
}

func newConfigGlobalPackageCmd(configPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "global-package",
		Short: "Manage saved global packages",
	}

	cmd.AddCommand(newConfigGlobalPackageAddCmd(configPath))
	cmd.AddCommand(newConfigGlobalPackageListCmd(configPath))
	cmd.AddCommand(newConfigGlobalPackageRemoveCmd(configPath))

	return cmd
}

func newConfigGlobalPackageAddCmd(configPath *string) *cobra.Command {
	var packageFlags []string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add packages to the saved global package list",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(packageFlags) == 0 {
				return custom_errors.CreateInvalidInputErrorWithMessage("at least one package is required")
			}
			trimmedPackages, err := validation.NonEmptyStrings(packageFlags, "package values")
			if err != nil {
				return err
			}
			if lo.ContainsBy(trimmedPackages, func(pkg string) bool {
				return !validation.IsFullModulePath(pkg)
			}) {
				return custom_errors.CreateInvalidInputErrorWithMessage(
					"package values must be full module paths (for example: github.com/user/module)",
				)
			}
			packageFlags = trimmedPackages

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogInfoIfProduction("config global-package add: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			values.GlobalPackages = lo.Uniq(append(values.GlobalPackages, packageFlags...))
			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			return cmdutil.WriteLine(cmd.OutOrStdout(), "global packages saved")
		},
	}

	cmd.Flags().StringSliceVar(&packageFlags, "package", nil, "full module paths to add")

	return cmd
}

func newConfigGlobalPackageRemoveCmd(configPath *string) *cobra.Command {
	var packageFlags []string

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove packages from the saved global package list",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(packageFlags) == 0 {
				return custom_errors.CreateInvalidInputErrorWithMessage("at least one package is required")
			}
			trimmedPackages, err := validation.NonEmptyStrings(packageFlags, "package values")
			if err != nil {
				return err
			}
			packageFlags = trimmedPackages

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogInfoIfProduction("config global-package remove: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			removeSet := lo.SliceToMap(packageFlags, func(pkg string) (string, struct{}) {
				return pkg, struct{}{}
			})
			filtered := lo.Filter(values.GlobalPackages, func(pkg string, _ int) bool {
				_, found := removeSet[pkg]
				return !found
			})

			if len(filtered) == len(values.GlobalPackages) {
				return custom_errors.CreateInvalidInputErrorWithMessage("no matching global packages found")
			}

			values.GlobalPackages = filtered
			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			return cmdutil.WriteLine(cmd.OutOrStdout(), "global packages updated")
		},
	}

	cmd.Flags().StringSliceVar(&packageFlags, "package", nil, "full module paths to remove")

	return cmd
}

func newConfigGlobalPackageListCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List saved global packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogInfoIfProduction("config global-package list: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			if len(values.GlobalPackages) > 0 {
				return cmdutil.WriteLine(cmd.OutOrStdout(), strings.Join(values.GlobalPackages, "\n"))
			}

			return nil
		},
	}
}

func newConfigRemoveCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove the config file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			targetPath := strings.TrimSpace(*configPath)
			if targetPath == "" {
				return custom_errors.CreateInvalidInputErrorWithMessage("config path is required")
			}

			cmdutil.LogInfoIfProduction("config remove: removing %s", targetPath)
			if err := os.Remove(targetPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}

			return cmdutil.WriteLine(cmd.OutOrStdout(), "config file removed")
		},
	}
}
