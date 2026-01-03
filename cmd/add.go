package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/louiss0/cobra-cli-template/custom_errors"
	"github.com/louiss0/cobra-cli-template/internal/cmdutil"
	"github.com/louiss0/cobra-cli-template/internal/config"
	"github.com/louiss0/cobra-cli-template/internal/packagepath"
	"github.com/louiss0/cobra-cli-template/internal/runner"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func NewAddCmd(commandRunner runner.Runner, configPath *string) *cobra.Command {
	var siteFlag string
	var userFlag string
	var allowFull bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "add <package> [packages...]",
		Short: "Add Go module dependencies",
		Args: func(cmd *cobra.Command, args []string) error {
			argErr := cobra.MinimumNArgs(1)(cmd, args)
			if argErr != nil {
				return argErr
			}

			for _, input := range args {
				if strings.Contains(input, "@none") {
					return custom_errors.CreateInvalidInputErrorWithMessage("do not use @none with add; use remove instead")
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			values, err := config.Load(*configPath)
			if err != nil {
				return err
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

			cmdutil.LogInfoIfProduction("add: resolving module paths for %s", site)

			modulePaths := make([]string, 0, len(args))
			for _, input := range args {
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
	cmdutil.RegisterSiteCompletion(cmd, "site")

	return cmd
}
