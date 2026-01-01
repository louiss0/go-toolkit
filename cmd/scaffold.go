package cmd

import (
	"errors"
	"path/filepath"

	"github.com/louiss0/cobra-cli-template/custom_errors"
	"github.com/louiss0/cobra-cli-template/internal/config"
	"github.com/louiss0/cobra-cli-template/internal/packagepath"
	"github.com/louiss0/cobra-cli-template/internal/runner"
	"github.com/louiss0/cobra-cli-template/internal/scaffold"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func NewScaffoldCmd(commandRunner runner.Runner, configPath *string) *cobra.Command {
	var folderFlag string
	var writeReadme bool
	var initModule bool
	var siteFlag string
	var userFlag string
	var allowFull bool

	cmd := &cobra.Command{
		Use:   "scaffold <package name>",
		Short: "Create a package folder with an optional index file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageName := packagepath.NormalizePackageName(args[0])

			folder := lo.Ternary(folderFlag == "", packageName, folderFlag)
			writeIndex := folderFlag == ""

			folder = filepath.Clean(folder)
			if err := scaffold.Create(folder, scaffold.Options{
				PackageName: packageName,
				WriteIndex:  writeIndex,
				WriteReadme: writeReadme,
			}); err != nil {
				return err
			}

			if !initModule {
				return nil
			}

			values, err := loadConfigValues(*configPath)
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
			if err := validateSite(site, allowCustomSite); err != nil {
				return err
			}

			modulePath, err := packagepath.ResolveModulePath(packageName, site, user)
			if err != nil {
				return err
			}

			return commandRunner.Run("go", []string{"-C", folder, "mod", "init", modulePath}, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}

	cmd.Flags().StringVar(&folderFlag, "folder", "", "use a custom folder and skip the index file")
	cmd.Flags().BoolVar(&writeReadme, "readme", false, "add a README.md to the package")
	cmd.Flags().BoolVar(&initModule, "module", false, "initialize a go.mod for the package")
	cmd.Flags().StringVar(&userFlag, "user", "", "override the configured user")
	cmd.Flags().StringVar(&siteFlag, "site", "", "override the configured site")
	cmd.Flags().BoolVar(&allowFull, "full", false, "allow a custom module site")
	registerSiteCompletion(cmd, "site")

	return cmd
}
