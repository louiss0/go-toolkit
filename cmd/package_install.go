package cmd

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/huh"
	"github.com/louiss0/go-toolkit/internal/cmdutil"
	"github.com/louiss0/go-toolkit/internal/modindex/config"
	"github.com/louiss0/go-toolkit/internal/prompt"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func parsePackageList(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})

	packages := lo.FilterMap(parts, func(part string, _ int) (string, bool) {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			return "", false
		}
		return trimmed, true
	})
	if len(packages) == 0 {
		return nil
	}

	return packages
}

func resolveInstallPackages(values config.Values, packageFlags []string, presetFlags []string, promptPackages []string) ([]string, error) {
	presetPackages, err := config.ResolvePackagePresetPackages(values, presetFlags)
	if err != nil {
		return nil, err
	}

	packages := make([]string, 0, len(packageFlags)+len(presetPackages)+len(promptPackages))
	packages = append(packages, packageFlags...)
	packages = append(packages, presetPackages...)
	packages = append(packages, promptPackages...)

	return lo.Uniq(packages), nil
}

const (
	packageProviderUseDefault = "use-default"
	packageProviderEdit       = "edit"
)

func assurePackageProviders(cmd *cobra.Command, runner prompt.Runner, values config.Values, defaultSite string, packages []string) ([]string, error) {
	if !values.AssureProviders {
		return packages, nil
	}

	shortIndexes := lo.FilterMap(packages, func(packageName string, index int) (int, bool) {
		return index, isShortPackageInput(packageName)
	})
	if len(shortIndexes) == 0 {
		return packages, nil
	}

	assuranceChoice, err := runner.Select(cmd, prompt.Select{
		Title: "Package providers",
		Options: []prompt.Option{
			{Label: fmt.Sprintf("Use %s for all", providerLabel(defaultSite)), Value: packageProviderUseDefault},
			{Label: "Edit providers", Value: packageProviderEdit},
		},
	})
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, err
		}
		return nil, err
	}
	if assuranceChoice == packageProviderUseDefault {
		return packages, nil
	}

	updatedPackages := append([]string{}, packages...)
	for _, packageIndex := range shortIndexes {
		providerSite, err := promptPackageProvider(cmd, runner, defaultSite, updatedPackages[packageIndex])
		if err != nil {
			return nil, err
		}

		updatedPackages[packageIndex] = providerSite + "/" + updatedPackages[packageIndex]
	}

	return updatedPackages, nil
}

func promptPackageProvider(cmd *cobra.Command, runner prompt.Runner, defaultSite string, packageName string) (string, error) {
	providerChoice, err := runner.Select(cmd, prompt.Select{
		Title:   fmt.Sprintf("Provider for %s", packageName),
		Options: buildPackageProviderOptions(defaultSite),
	})
	if err != nil {
		return "", err
	}

	if providerChoice != providerCustom {
		return providerChoice, nil
	}

	customSite, err := runner.Input(cmd, prompt.Input{
		Title:       "Custom provider",
		Placeholder: defaultSite,
		Validate: func(value string) error {
			if strings.TrimSpace(value) == "" {
				return errors.New("provider is required")
			}
			return cmdutil.ValidateSite(value, true)
		},
	})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(customSite), nil
}

func buildPackageProviderOptions(defaultSite string) []prompt.Option {
	options := lo.Map(config.KnownSites(), func(site string, _ int) prompt.Option {
		label := providerLabel(site)
		if site == defaultSite {
			label += " (Default)"
		}
		return prompt.Option{Label: label, Value: site}
	})

	if !config.IsKnownSite(defaultSite) {
		options = append([]prompt.Option{{
			Label: providerLabel(defaultSite) + " (Default)",
			Value: defaultSite,
		}}, options...)
	}

	options = append(options, prompt.Option{Label: "Custom", Value: providerCustom})
	return options
}

func providerLabel(site string) string {
	if knownLabel, ok := config.KnownSiteLabel(site); ok {
		return knownLabel
	}

	return site
}

func isShortPackageInput(value string) bool {
	parts := strings.Split(strings.TrimSpace(value), "/")
	if len(parts) != 2 {
		return false
	}

	return !strings.Contains(parts[0], ".")
}
