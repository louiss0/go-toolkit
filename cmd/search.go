package cmd

import (
	"fmt"
	"regexp"

	"github.com/louiss0/cobra-cli-template/custom_errors"
	"github.com/louiss0/cobra-cli-template/internal/cmdutil"
	"github.com/louiss0/cobra-cli-template/internal/search"
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

			standardQuery := `^[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+$`
			prefixedQuery := `^[a-zA-Z0-9-]+\.[a-zA-Z0-9-]+/(?:[a-zA-Z0-9_-]+/)+[a-zA-Z0-9_-]+$`

			standardMatch, _ := regexp.MatchString(standardQuery, query)
			prefixedMatch, _ := regexp.MatchString(prefixedQuery, query)

			if !standardMatch && !prefixedMatch {
				return custom_errors.CreateInvalidArgumentErrorWithMessage("query must be in the form scope/package")
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
				fmt.Fprintln(cmd.OutOrStdout(), version)
			}

			return nil
		},
	}

	return cmd
}
