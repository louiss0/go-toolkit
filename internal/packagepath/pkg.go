package packagepath

import (
	"errors"
	"strings"

	"github.com/louiss0/go-toolkit/custom_errors"
	"github.com/louiss0/go-toolkit/validation"
	"github.com/samber/lo"
)

var ErrMissingUser = errors.New("missing registered user")

func NormalizePackageName(value string) string {
	parts := strings.Fields(value)
	return strings.Join(parts, "_")
}

func ResolveModulePath(input string, site string, user string) (string, error) {
	trimmed, err := validation.RequiredString(input, "module path")
	if err != nil {
		return "", err
	}

	parts := lo.Map(strings.Split(trimmed, "/"), func(part string, _ int) string {
		return strings.TrimSpace(part)
	})
	if lo.ContainsBy(parts, func(part string) bool {
		return part == ""
	}) {
		return "", custom_errors.CreateInvalidInputErrorWithMessage("module path must not be empty")
	}

	if len(parts) >= 3 {
		if strings.Contains(parts[0], ".") {
			return strings.Join(parts, "/"), nil
		}
		if len(parts) == 3 && validation.IsShortPackagePath(trimmed) {
			if !validation.IsValidSite(site) {
				return "", custom_errors.CreateInvalidInputErrorWithMessage("site must be in the form sitename.domain")
			}
			return joinPath(site, parts[0], parts[1], parts[2]), nil
		}
		if len(parts) == 3 {
			return strings.Join(parts, "/"), nil
		}
		return "", custom_errors.CreateInvalidInputErrorWithMessage("module path must have 1 to 3 segments")
	}

	if len(parts) == 2 {
		if !validation.IsValidSite(site) {
			return "", custom_errors.CreateInvalidInputErrorWithMessage("site must be in the form sitename.domain")
		}
		return joinPath(site, parts[0], parts[1]), nil
	}

	if len(parts) == 1 {
		if user == "" {
			return "", ErrMissingUser
		}
		if !validation.IsValidSite(site) {
			return "", custom_errors.CreateInvalidInputErrorWithMessage("site must be in the form sitename.domain")
		}
		return joinPath(site, user, parts[0]), nil
	}

	return "", custom_errors.CreateInvalidInputErrorWithMessage("module path must have 1 to 3 segments")
}

func joinPath(site string, parts ...string) string {
	if site == "" {
		site = "github.com"
	}

	segments := append([]string{site}, parts...)
	return strings.Join(segments, "/")
}
