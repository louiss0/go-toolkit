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
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func NewUninstallCmd(commandRunner runner.Runner, promptRunner prompt.Runner, configPath *string) *cobra.Command {
	siteFlag := custom_flags.NewEmptyStringFlag("site")
	userFlag := custom_flags.NewEmptyStringFlag("user")
	var allowFull bool
	var dryRun bool
	var presetFlags []string
	var packageFlags []string

	cmd := &cobra.Command{
		Use:   "uninstall [package] [packages...]",
		Short: "Uninstall Go binaries globally and remove them from the global package list",
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

			cmdutil.LogInfoIfProduction("uninstall: resolving module paths for %s", site)
			modulePaths, err := resolveModulePaths(targetPackages, site, user)
			if err != nil {
				return err
			}

			basePaths := lo.Map(modulePaths, func(modulePath string, _ int) string {
				return strings.Split(modulePath, "@")[0]
			})

			if dryRun {
				cmdutil.LogInfoIfProduction("uninstall: dry run output")
				lines := lo.Map(basePaths, func(modulePath string, _ int) string {
					return "go clean -i " + modulePath
				})
				return cmdutil.WriteLine(cmd.OutOrStdout(), strings.Join(lines, "\n"))
			}

			cmdutil.LogInfoIfProduction("uninstall: executing go clean -i")
			for _, modulePath := range basePaths {
				if err := commandRunner.Run(cmd, "go", "clean", "-i", modulePath); err != nil {
					return err
				}
			}

			removeSet := lo.SliceToMap(basePaths, func(modulePath string) (string, struct{}) {
				return modulePath, struct{}{}
			})
			values.GlobalPackages = lo.Filter(values.GlobalPackages, func(modulePath string, _ int) bool {
				_, shouldRemove := removeSet[modulePath]
				return !shouldRemove
			})
			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			return cmdutil.WriteLine(cmd.OutOrStdout(), "uninstalled and removed from global packages")
		},
	}

	cmd.Flags().Var(&userFlag, "user", "override the configured user")
	cmd.Flags().Var(&siteFlag, "site", "override the configured site")
	cmd.Flags().BoolVar(&allowFull, "full", false, "allow a custom module site")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the go command without running it")
	cmd.Flags().StringSliceVar(&packageFlags, "package", nil, "package preset entries or module paths to uninstall")
	cmd.Flags().StringSliceVar(&presetFlags, "preset", nil, "package preset names to uninstall")
	cmdutil.RegisterSiteCompletion(cmd, "site")

	return cmd
}
