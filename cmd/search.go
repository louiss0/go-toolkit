package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/louiss0/cobra-cli-template/internal/config"
	"github.com/louiss0/cobra-cli-template/internal/modindex"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func NewSearchCmd(fetcher modindex.Fetcher, configPath *string) *cobra.Command {
	var siteFlag string
	var since string
	var sinceDays int
	var sinceHours int
	var limit int
	var includeAll bool
	var allowFull bool
	var showDetails bool
	var maxPages int
	var maxResults int
	var useJSON bool

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

			if since != "" && (sinceDays > 0 || sinceHours > 0) {
				return fmt.Errorf("since cannot be combined with since-days or since-hours")
			}

			if sinceDays < 0 || sinceHours < 0 {
				return fmt.Errorf("since-days and since-hours must be non-negative")
			}

			if since == "" && sinceDays > 0 && sinceHours > 0 {
				return fmt.Errorf("since-days and since-hours cannot both be set")
			}

			if since == "" && sinceDays > 0 {
				since = time.Now().UTC().Add(-time.Duration(sinceDays) * 24 * time.Hour).Format(time.RFC3339)
			}

			if since == "" && sinceHours > 0 {
				since = time.Now().UTC().Add(-time.Duration(sinceHours) * time.Hour).Format(time.RFC3339)
			}

			if since != "" {
				if _, err := time.Parse(time.RFC3339, since); err != nil {
					return fmt.Errorf("since must be RFC3339: %w", err)
				}
			}

			if maxPages < 0 || maxResults < 0 {
				return fmt.Errorf("max-pages and max-results must be non-negative")
			}

			if maxPages > 0 && limit <= 0 {
				return fmt.Errorf("max-pages requires a positive limit")
			}

			site := config.ResolveSite(siteFlag, values)
			allowCustomSite := allowFull || (siteFlag == "" && values.Site != "")
			if err := validateSite(site, allowCustomSite); err != nil {
				return err
			}

			entries, err := modindex.FetchAll(cmd.Context(), fetcher, modindex.Request{
				Since:      since,
				Limit:      limit,
				IncludeAll: includeAll,
				MaxPages:   maxPages,
				MaxResults: maxResults,
			})
			if err != nil {
				return err
			}

			query := strings.TrimSpace(args[0])
			useSiteFilter := !strings.Contains(query, ".")

			filtered := modindex.FilterEntries(entries, query, site, useSiteFilter)
			outputEntries := filtered
			if showDetails {
				outputEntries = filtered
			} else {
				outputEntries = lo.UniqBy(filtered, func(entry modindex.Entry) string {
					return entry.Path
				})
			}

			if useJSON {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				return encoder.Encode(outputEntries)
			}

			if useSiteFilter && site != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "filtering to %s\n", site)
			}

			if showDetails {
				for _, entry := range outputEntries {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", entry.Path, entry.Version, entry.Timestamp)
				}
				return nil
			}

			for _, entry := range outputEntries {
				fmt.Fprintln(cmd.OutOrStdout(), entry.Path)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&siteFlag, "site", "", "restrict results to a module site")
	cmd.Flags().StringVar(&since, "since", "", "oldest timestamp (RFC3339) to include")
	cmd.Flags().IntVar(&sinceDays, "since-days", 0, "days since now to include")
	cmd.Flags().IntVar(&sinceHours, "since-hours", 0, "hours since now to include")
	cmd.Flags().IntVar(&limit, "limit", 200, "max results to request from the index")
	cmd.Flags().IntVar(&maxPages, "max-pages", 0, "max pages to request from the index")
	cmd.Flags().IntVar(&maxResults, "max-results", 0, "max results to return")
	cmd.Flags().BoolVar(&includeAll, "include-all", false, "include uncached module versions")
	cmd.Flags().BoolVar(&allowFull, "full", false, "allow a custom module site")
	cmd.Flags().BoolVar(&showDetails, "details", false, "show path, version, and timestamp")
	cmd.Flags().BoolVar(&useJSON, "json", false, "output results as JSON")
	registerSiteCompletion(cmd, "site")

	return cmd
}
