package packagepath

import (
	"errors"
	"strings"

	"github.com/louiss0/cobra-cli-template/custom_errors"
)

var ErrMissingUser = errors.New("missing registered user")

func NormalizePackageName(value string) string {
	parts := strings.Fields(value)
	return strings.Join(parts, "_")
}

func ResolveModulePath(input string, site string, user string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", errors.New("module path is required")
	}

	parts := strings.Split(trimmed, "/")
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
		if parts[i] == "" {
			return "", custom_errors.CreateInvalidInputErrorWithMessage("module path must not be empty")
		}
	}

	if len(parts) >= 3 {
		if strings.Contains(parts[0], ".") {
			return strings.Join(parts, "/"), nil
		}
		if len(parts) == 3 {
			return strings.Join(parts, "/"), nil
		}
		return "", custom_errors.CreateInvalidInputErrorWithMessage("module path must have 1 to 3 segments")
	}

	if len(parts) == 2 {
		if !isValidSite(site) {
			return "", custom_errors.CreateInvalidInputErrorWithMessage("site must be in the form sitename.domain")
		}
		return joinPath(site, parts[0], parts[1]), nil
	}

	if len(parts) == 1 {
		if user == "" {
			return "", ErrMissingUser
		}
		if !isValidSite(site) {
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

func isValidSite(site string) bool {
	trimmed := strings.TrimSpace(site)
	if trimmed == "" {
		return false
	}

	if strings.Contains(trimmed, " ") {
		return false
	}

	return strings.Contains(trimmed, ".") && !strings.HasPrefix(trimmed, ".") && !strings.HasSuffix(trimmed, ".")
}
