package cmd

import (
	"github.com/louiss0/go-toolkit/custom_errors"
	"github.com/louiss0/go-toolkit/internal/cmdutil"
	"github.com/louiss0/go-toolkit/internal/search"
	"github.com/louiss0/go-toolkit/validation"
	"github.com/spf13/cobra"
)

func NewSearchCmd() *cobra.Command {

	var modulePath string

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "List module versions from the Go proxy",
		Args: func(cmd *cobra.Command, args []string) error {
			argErr := cobra.ExactArgs(1)(cmd, args)

			if argErr != nil {
				return argErr
			}

			query := args[0]
			if !validation.IsShortPackagePath(query) && !validation.IsFullModulePath(query) {
				return custom_errors.CreateInvalidArgumentErrorWithMessage(
					"query must be in the form scope/package or scope/package/vN",
				)
			}

			modulePath = search.ResolveModulePath(query)

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogInfoIfProduction("search: fetching module versions for %s", modulePath)
			versions, err := search.FetchModuleVersions(cmd.Context(), modulePath)

			if err != nil {
				return err
			}

			for _, version := range versions {
				if err := cmdutil.WriteLine(cmd.OutOrStdout(), version); err != nil {
					return err
				}
			}

			return nil
		},
	}

	return cmd
}
