package cmd

import (
	"errors"
	"strings"

	"github.com/louiss0/go-toolkit/custom_errors"
	"github.com/louiss0/go-toolkit/custom_flags"
	"github.com/louiss0/go-toolkit/internal/cmdutil"
	"github.com/louiss0/go-toolkit/internal/modindex/config"
	"github.com/louiss0/go-toolkit/internal/packagepath"
	"github.com/louiss0/go-toolkit/internal/runner"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func NewRemoveCmd(commandRunner runner.Runner, configPath *string) *cobra.Command {
	siteFlag := custom_flags.NewEmptyStringFlag("site")
	userFlag := custom_flags.NewEmptyStringFlag("user")
	var allowFull bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "remove <package> [packages...]",
		Short: "Remove a Go module dependency",
		Args: func(cmd *cobra.Command, args []string) error {
			argErr := cobra.MinimumNArgs(1)(cmd, args)
			if argErr != nil {
				return argErr
			}

			containsNoneTag := lo.ContainsBy(args, func(input string) bool {
				return strings.Contains(input, "@none")
			})
			if containsNoneTag {
				return custom_errors.CreateInvalidInputErrorWithMessage("@none is added automatically; omit it from remove")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			values, err := config.Load(*configPath)
			if err != nil {
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

			cmdutil.LogInfoIfProduction("remove: resolving module paths for %s", site)

			modulePaths := make([]string, 0, len(args))
			for _, input := range args {
				modulePath, err := packagepath.ResolveModulePath(input, site, user)
				if err != nil {
					return err
				}

				modulePath = strings.Split(modulePath, "@")[0]
				modulePaths = append(modulePaths, modulePath+"@none")
			}

			uniqueModules := lo.Uniq(modulePaths)
			if dryRun {
				cmdutil.LogInfoIfProduction("remove: dry run output")
				return cmdutil.WriteLine(
					cmd.OutOrStdout(),
					"go "+strings.Join(append([]string{"get"}, uniqueModules...), " "),
				)
			}

			cmdutil.LogInfoIfProduction("remove: executing go get")
			if err := commandRunner.Run(cmd, "go", append([]string{"get"}, uniqueModules...)...); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().Var(&userFlag, "user", "override the configured user")
	cmd.Flags().Var(&siteFlag, "site", "override the configured site")
	cmd.Flags().BoolVar(&allowFull, "full", false, "allow a custom module site")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the go command without running it")
	cmdutil.RegisterSiteCompletion(cmd, "site")

	return cmd
}
