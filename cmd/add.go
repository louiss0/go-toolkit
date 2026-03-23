package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/louiss0/go-toolkit/custom_errors"
	"github.com/louiss0/go-toolkit/internal/cmdutil"
	"github.com/louiss0/go-toolkit/internal/modindex/config"
	"github.com/louiss0/go-toolkit/internal/packagepath"
	"github.com/louiss0/go-toolkit/internal/prompt"
	"github.com/louiss0/go-toolkit/internal/runner"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func NewAddCmd(commandRunner runner.Runner, promptRunner prompt.Runner, configPath *string) *cobra.Command {
	var siteFlag string
	var userFlag string
	var allowFull bool
	var dryRun bool
	var presetFlags []string
	var packageFlags []string

	cmd := &cobra.Command{
		Use:   "add [package] [packages...]",
		Short: "Add Go module dependencies",
		Args: func(cmd *cobra.Command, args []string) error {
			containsNoneTag := lo.ContainsBy(args, func(input string) bool {
				return strings.Contains(input, "@none")
			})
			if containsNoneTag {
				return custom_errors.CreateInvalidInputErrorWithMessage("do not use @none with add; use remove instead")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}
			promptPackages := []string(nil)
			if len(args) == 0 && len(packageFlags) == 0 && len(presetFlags) == 0 {
				inputs, err := promptAddPackages(cmd, promptRunner)
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

			site := config.ResolveSite(siteFlag, values)
			user, err := config.ResolveUser(userFlag, values, site)
			if err != nil {
				if errors.Is(err, config.ErrMissingUser) {
					return custom_errors.CreateInvalidInputErrorWithMessage("missing user; run go-toolkit config set-user <user>")
				}
				return err
			}

			allowCustomSite := allowFull || (siteFlag == "" && values.Site != "")
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

			cmdutil.LogInfoIfProduction("add: resolving module paths for %s", site)

			modulePaths := make([]string, 0, len(targetPackages))
			for _, input := range targetPackages {
				modulePath, err := packagepath.ResolveModulePath(input, site, user)
				if err != nil {
					return err
				}

				modulePaths = append(modulePaths, modulePath)
			}

			uniqueModules := lo.Uniq(modulePaths)
			if dryRun {
				cmdutil.LogInfoIfProduction("add: dry run output")
				fmt.Fprintln(cmd.OutOrStdout(), "go "+strings.Join(append([]string{"get"}, uniqueModules...), " "))
				return nil
			}

			cmdutil.LogInfoIfProduction("add: executing go get")
			if err := commandRunner.Run(cmd, "go", append([]string{"get"}, uniqueModules...)...); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&userFlag, "user", "", "override the configured user")
	cmd.Flags().StringVar(&siteFlag, "site", "", "override the configured site")
	cmd.Flags().BoolVar(&allowFull, "full", false, "allow a custom module site")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the go command without running it")
	cmd.Flags().StringSliceVar(&packageFlags, "package", nil, "package preset entries or module paths to add")
	cmd.Flags().StringSliceVar(&presetFlags, "preset", nil, "package preset names to add")
	cmdutil.RegisterSiteCompletion(cmd, "site")

	return cmd
}

func promptAddPackages(cmd *cobra.Command, runner prompt.Runner) ([]string, error) {
	packageInput, err := runner.Input(cmd, prompt.Input{
		Title:       "Packages to add",
		Description: "Space or comma separated package paths; presets can be used with --preset.",
		Placeholder: "samber/lo, stretchr/testify",
		Validate: func(value string) error {
			if len(parsePackageList(value)) == 0 {
				return errors.New("at least one package is required")
			}
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	return parsePackageList(packageInput), nil
}
