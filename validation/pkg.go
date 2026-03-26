package validation

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/kaptinlin/gozod"
	"github.com/louiss0/go-toolkit/custom_errors"
	"github.com/samber/lo"
)

var (
	requiredStringSchema = gozod.String().Regex(regexp.MustCompile(`.+`))
	siteSchema           = gozod.String().Regex(regexp.MustCompile(`^[^\s.][^\s]*\.[^\s]*[^\s.]$`))
	booleanStringSchema  = gozod.String().Regex(regexp.MustCompile(`(?i)^(1|0|t|f|true|false)$`))
	versionSegmentSchema = regexp.MustCompile(`^v[0-9].*$`)
)

const shortPackageListFormatMessage = "username/package or username/package/vN"

func RequiredString(value string, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if _, err := requiredStringSchema.Parse(trimmed); err != nil {
		return "", custom_errors.CreateInvalidInputErrorWithMessage(field + " is required")
	}

	return trimmed, nil
}

func NonEmptyStrings(values []string, field string) ([]string, error) {
	trimmedValues := lo.Map(values, func(value string, _ int) string {
		return strings.TrimSpace(value)
	})
	if lo.ContainsBy(trimmedValues, func(value string) bool {
		return value == ""
	}) {
		return nil, custom_errors.CreateInvalidInputErrorWithMessage(field + " must not be empty")
	}

	return trimmedValues, nil
}

func ParseBool(value string, field string) (bool, error) {
	trimmed := strings.TrimSpace(value)
	if _, err := booleanStringSchema.Parse(trimmed); err != nil {
		return false, custom_errors.CreateInvalidInputErrorWithMessage(field + " must be true or false")
	}

	parsedValue, err := strconv.ParseBool(trimmed)
	if err != nil {
		return false, custom_errors.CreateInvalidInputErrorWithMessage(field + " must be true or false")
	}

	return parsedValue, nil
}

func IsValidSite(site string) bool {
	trimmed := strings.TrimSpace(site)
	if trimmed == "" {
		return false
	}

	if _, err := siteSchema.Parse(trimmed); err != nil {
		return false
	}

	return true
}

func ValidateSite(site string, allowFull bool, knownSites []string) error {
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

	if _, err := gozod.Enum(knownSites...).Parse(trimmed); err == nil {
		return nil
	}

	known := strings.Join(knownSites, ", ")
	return custom_errors.CreateInvalidInputErrorWithMessage(
		"unsupported site " + trimmed + " (known: " + known + "). use --full to allow custom sites",
	)
}

func ParseShortPackageList(value string, field string) ([]string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	packages := strings.Fields(trimmed)
	if lo.ContainsBy(packages, func(packageName string) bool {
		return !IsShortPackagePath(packageName)
	}) {
		return nil, custom_errors.CreateInvalidInputErrorWithMessage(
			field + " must use space-separated " + shortPackageListFormatMessage + " entries",
		)
	}

	return packages, nil
}

func RequiredShortPackageList(value string, field string) ([]string, error) {
	packages, err := ParseShortPackageList(value, field)
	if err != nil {
		return nil, err
	}

	if len(packages) == 0 {
		return nil, custom_errors.CreateInvalidInputErrorWithMessage(
			field + " must use space-separated " + shortPackageListFormatMessage + " entries",
		)
	}

	return packages, nil
}

func IsShortPackagePath(value string) bool {
	parts := strings.Split(strings.TrimSpace(value), "/")
	if len(parts) != 2 && len(parts) != 3 {
		return false
	}

	if strings.Contains(parts[0], ".") {
		return false
	}
	if len(parts) == 3 && !versionSegmentSchema.MatchString(parts[2]) {
		return false
	}

	return lo.EveryBy(parts, func(part string) bool {
		return part != "" && !strings.ContainsAny(part, ", \t\r\n")
	})
}

func IsFullModulePath(value string) bool {
	trimmed := strings.TrimSpace(value)
	parts := strings.Split(trimmed, "/")
	if len(parts) < 2 {
		return false
	}

	if !strings.Contains(parts[0], ".") {
		return false
	}

	return lo.EveryBy(parts, func(part string) bool {
		return part != "" && !strings.ContainsAny(part, ", \t\r\n")
	})
}
