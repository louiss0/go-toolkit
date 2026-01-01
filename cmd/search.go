package cmd

import (
	"fmt"
	"strings"

	"github.com/louiss0/cobra-cli-template/internal/config"
	"github.com/louiss0/cobra-cli-template/internal/modindex"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func NewSearchCmd(fetcher modindex.Fetcher, configPath *string) *cobra.Command {
	var siteFlag string
	var since string
	var limit int
	var includeAll bool
	var allowFull bool
	var showDetails bool

	if fetcher == nil {
		fetcher = modindex.HTTPFetcher{}
	}

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search the Go module index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			values, err := loadConfigValues(*configPath)
			if err != nil {
				return err
			}

			site := config.ResolveSite(siteFlag, values)
			allowCustomSite := allowFull || (siteFlag == "" && values.Site != "")
			if err := validateSite(site, allowCustomSite); err != nil {
				return err
			}

			entries, err := fetcher.Fetch(cmd.Context(), modindex.Request{
				Since:      since,
				Limit:      limit,
				IncludeAll: includeAll,
			})
			if err != nil {
				return err
			}

			query := strings.TrimSpace(args[0])
			useSiteFilter := !strings.Contains(query, ".")

			filtered := modindex.FilterEntries(entries, query, site, useSiteFilter)
			if showDetails {
				for _, entry := range filtered {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", entry.Path, entry.Version, entry.Timestamp)
				}
				return nil
			}

			unique := lo.UniqBy(filtered, func(entry modindex.Entry) string {
				return entry.Path
			})
			for _, entry := range unique {
				fmt.Fprintln(cmd.OutOrStdout(), entry.Path)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&siteFlag, "site", "", "restrict results to a module site")
	cmd.Flags().StringVar(&since, "since", "", "oldest timestamp (RFC3339) to include")
	cmd.Flags().IntVar(&limit, "limit", 200, "max results to request from the index")
	cmd.Flags().BoolVar(&includeAll, "include-all", false, "include uncached module versions")
	cmd.Flags().BoolVar(&allowFull, "full", false, "allow a custom module site")
	cmd.Flags().BoolVar(&showDetails, "details", false, "show path, version, and timestamp")
	registerSiteCompletion(cmd, "site")

	return cmd
}
