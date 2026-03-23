package config

import (
	"fmt"
	"slices"
	"strings"
)

func KnownPackagePresetNames(values Values) []string {
	names := make([]string, 0, len(values.PackagePresets))
	for name := range values.PackagePresets {
		names = append(names, name)
	}
	slices.Sort(names)

	return names
}

func ResolvePackagePresetPackages(values Values, presetNames []string) ([]string, error) {
	packages := make([]string, 0)
	for _, presetName := range presetNames {
		presetPackages, ok := values.PackagePresets[presetName]
		if !ok {
			return nil, fmt.Errorf("unknown package preset: %s", presetName)
		}

		packages = append(packages, presetPackages...)
	}

	seenPackages := map[string]struct{}{}
	uniquePackages := make([]string, 0, len(packages))
	for _, packageName := range packages {
		if _, seen := seenPackages[packageName]; seen {
			continue
		}

		seenPackages[packageName] = struct{}{}
		uniquePackages = append(uniquePackages, packageName)
	}

	return uniquePackages, nil
}

func validatePackagePresets(presets map[string][]string) error {
	for name, packages := range presets {
		if name == "" {
			return fmt.Errorf("invalid config values: package preset name is required")
		}
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("invalid config values: package preset name is required")
		}
		if len(packages) == 0 {
			return fmt.Errorf("invalid config values: package preset %s must include at least one package", name)
		}
		for _, packageName := range packages {
			if strings.TrimSpace(packageName) == "" {
				return fmt.Errorf("invalid config values: package preset %s contains an empty package", name)
			}
		}
	}

	return nil
}
