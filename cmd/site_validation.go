package cmd

import (
	"strings"

	"github.com/louiss0/cobra-cli-template/custom_errors"
	"github.com/louiss0/cobra-cli-template/internal/config"
	"github.com/spf13/cobra"
)

func validateSite(site string, allowFull bool) error {
	if site == "" {
		return nil
	}

	if !config.IsValidSite(site) {
		return custom_errors.CreateInvalidInputErrorWithMessage("site must be in the form sitename.domain")
	}

	if allowFull || config.IsKnownSite(site) {
		return nil
	}

	known := strings.Join(config.KnownSites(), ", ")
	return custom_errors.CreateInvalidInputErrorWithMessage(
		"unsupported site " + site + " (known: " + known + "). use --full to allow custom sites",
	)
}

func registerSiteCompletion(cmd *cobra.Command, flagName string) {
	_ = cmd.RegisterFlagCompletionFunc(flagName, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return config.KnownSites(), cobra.ShellCompDirectiveNoFileComp
	})
}
