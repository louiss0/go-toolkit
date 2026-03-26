package cmd

import (
	"errors"
	"strings"

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

func NewInstallCmd(commandRunner runner.Runner, promptRunner prompt.Runner, configPath *string) *cobra.Command {
	siteFlag := custom_flags.NewEmptyStringFlag("site")
	userFlag := custom_flags.NewEmptyStringFlag("user")
	var allowFull bool
	var dryRun bool
	var presetFlags []string
	var packageFlags []string

	cmd := &cobra.Command{
		Use:   "install [package] [packages...]",
		Short: "Install Go binaries globally and save them to the global package list",
		Args: func(cmd *cobra.Command, args []string) error {
			return validateInstallInputs(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}
			promptPackages := []string(nil)
			if len(args) == 0 && len(packageFlags) == 0 && len(presetFlags) == 0 {
				inputs, err := promptInstallPackages(cmd, promptRunner)
				if err != nil {
					if errors.Is(err, huh.ErrUserAborted) {
						return nil
					}
					return err
				}
				promptPackages = inputs
			}

			installPackages, err := resolveInstallPackages(values, packageFlags, presetFlags, promptPackages)
			if err != nil {
				return err
			}
			targetPackages := append([]string{}, args...)
			targetPackages = append(targetPackages, installPackages...)
			if len(targetPackages) == 0 {
				return custom_errors.CreateInvalidInputErrorWithMessage("at least one package or preset is required")
			}
			if err := validateInstallInputs(targetPackages); err != nil {
				return err
			}

			site := config.ResolveSite(siteFlag.String(), values)
			user, err := config.ResolveUser(userFlag.String(), values, site)
			if err != nil {
				if errors.Is(err, config.ErrMissingUser) {
					return custom_errors.CreateInvalidInputErrorWithMessage("missing user; run go-toolkit config set-user <user>")
				}
				return err
			}

			allowCustomSite := allowFull || (siteFlag.String() == "" && values.Site != "")
			if err := cmdutil.ValidateSite(site, allowCustomSite); err != nil {
				return err
			}

			targetPackages, err = assurePackageProviders(cmd, promptRunner, values, site, targetPackages)
			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					return nil
				}
				return err
			}

			cmdutil.LogInfoIfProduction("install: resolving module paths for %s", site)
			uniqueModules, err := resolveModulePaths(targetPackages, site, user)
			if err != nil {
				return err
			}

			// Strip any existing version suffix and append @latest
			installArgs := lo.Map(uniqueModules, func(mod string, _ int) string {
				base := strings.Split(mod, "@")[0]
				return base + "@latest"
			})

			if dryRun {
				cmdutil.LogInfoIfProduction("install: dry run output")
				lines := lo.Map(installArgs, func(arg string, _ int) string {
					return "go install " + arg
				})
				return cmdutil.WriteLine(cmd.OutOrStdout(), strings.Join(lines, "\n"))
			}

			cmdutil.LogInfoIfProduction("install: executing go install")
			for _, arg := range installArgs {
				if err := commandRunner.Run(cmd, "go", "install", arg); err != nil {
					return err
				}
			}

			// Save installed packages to the global packages list
			basePaths := lo.Map(uniqueModules, func(mod string, _ int) string {
				return strings.Split(mod, "@")[0]
			})
			values.GlobalPackages = lo.Uniq(append(values.GlobalPackages, basePaths...))
			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			return cmdutil.WriteLine(cmd.OutOrStdout(), "installed and saved to global packages")
		},
	}

	cmd.Flags().Var(&userFlag, "user", "override the configured user")
	cmd.Flags().Var(&siteFlag, "site", "override the configured site")
	cmd.Flags().BoolVar(&allowFull, "full", false, "allow a custom module site")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the go command without running it")
	cmd.Flags().StringSliceVar(&packageFlags, "package", nil, "package preset entries or module paths to install")
	cmd.Flags().StringSliceVar(&presetFlags, "preset", nil, "package preset names to install")
	cmdutil.RegisterSiteCompletion(cmd, "site")

	return cmd
}

func promptInstallPackages(cmd *cobra.Command, runner prompt.Runner) ([]string, error) {
	packageInput, err := runner.Input(cmd, prompt.Input{
		Title:       "Packages to install globally",
		Description: "Use space-separated username/package or username/package/vN entries; presets can be used with --preset.",
		Placeholder: "samber/lo onsi/ginkgo/v2",
		Validate: func(value string) error {
			_, err := validation.RequiredShortPackageList(value, "packages to install")
			return err
		},
	})
	if err != nil {
		return nil, err
	}

	return validation.RequiredShortPackageList(packageInput, "packages to install")
}

func validateInstallInputs(inputs []string) error {
	containsNoneTag := lo.ContainsBy(inputs, func(input string) bool {
		return strings.Contains(input, "@none")
	})
	if containsNoneTag {
		return custom_errors.CreateInvalidInputErrorWithMessage("@none is not valid for install")
	}

	return nil
}
