package cmd

import (
	"strings"
	"unicode"

	"github.com/louiss0/go-toolkit/internal/modindex/config"
	"github.com/samber/lo"
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
