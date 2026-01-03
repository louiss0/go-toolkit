package cmd

import (
	"errors"

	"github.com/louiss0/cobra-cli-template/custom_errors"
	"github.com/louiss0/cobra-cli-template/internal/cmdutil"
	"github.com/louiss0/cobra-cli-template/internal/config"
	"github.com/louiss0/cobra-cli-template/internal/packagepath"
	"github.com/louiss0/cobra-cli-template/internal/runner"
	"github.com/spf13/cobra"
)

func NewInitCmd(commandRunner runner.Runner, configPath *string) *cobra.Command {
	var siteFlag string
	var userFlag string
	var allowFull bool

	cmd := &cobra.Command{
		Use:   "init <package>",
		Short: "Initialize a Go module with a short package name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogInfoIfProduction("init: loading config")
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

			cmdutil.LogInfoIfProduction("init: resolving module path for %s", site)
			modulePath, err := packagepath.ResolveModulePath(args[0], site, user)
			if err != nil {
				return err
			}

			cmdutil.LogInfoIfProduction("init: running go mod init")
			if err := commandRunner.Run(cmd, "go", "mod", "init", modulePath); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&userFlag, "user", "", "override the configured user")
	cmd.Flags().StringVar(&siteFlag, "site", "", "override the configured site")
	cmd.Flags().BoolVar(&allowFull, "full", false, "allow a custom module site")
	cmdutil.RegisterSiteCompletion(cmd, "site")

	return cmd
}
