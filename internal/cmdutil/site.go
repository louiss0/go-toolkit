package cmdutil

import (
	"regexp"
	"strings"

	"github.com/kaptinlin/gozod"
	"github.com/louiss0/cobra-cli-template/custom_errors"
	"github.com/louiss0/cobra-cli-template/internal/config"
	"github.com/spf13/cobra"
)

var siteSchema = gozod.String().Regex(regexp.MustCompile(`^[^\s.][^\s]*\.[^\s]*[^\s.]$`))

func ValidateSite(site string, allowFull bool) error {
	trimmed := strings.TrimSpace(site)
	if trimmed == "" {
		return nil
	}

	if _, err := siteSchema.Parse(trimmed); err != nil {
		return custom_errors.CreateInvalidInputErrorWithMessage("site must be in the form sitename.domain")
	}

	if allowFull {
		return nil
	}

	if _, err := gozod.Enum(config.KnownSites()...).Parse(trimmed); err == nil {
		return nil
	}

	known := strings.Join(config.KnownSites(), ", ")
	return custom_errors.CreateInvalidInputErrorWithMessage(
		"unsupported site " + trimmed + " (known: " + known + "). use --full to allow custom sites",
	)
}

func RegisterSiteCompletion(cmd *cobra.Command, flagName string) {
	_ = cmd.RegisterFlagCompletionFunc(flagName, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return config.KnownSites(), cobra.ShellCompDirectiveNoFileComp
	})
}
