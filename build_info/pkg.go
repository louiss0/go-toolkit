package build_info

import (
	"strings"
	"time"
)

type BuildInfo string

func (value BuildInfo) String() string {
	return string(value)
}

var (
	rawCLI_VERSION = "dev"
	rawBUILD_DATE  = "unknown"
)

var (
	CLI_VERSION BuildInfo
	BUILD_DATE  BuildInfo
)

func init() {
	CLI_VERSION = BuildInfo(normalizeVersion(rawCLI_VERSION))
	BUILD_DATE = BuildInfo(normalizeBuildDate(rawBUILD_DATE))
}

func Version() string {
	return CLI_VERSION.String()
}

func BuildDate() string {
	return BUILD_DATE.String()
}

func normalizeVersion(rawVersion string) string {
	return strings.TrimPrefix(rawVersion, "v")
}

func normalizeBuildDate(rawDate string) string {
	if rawDate == "" || rawDate == "unknown" {
		return "unknown"
	}

	parsedDate, err := time.Parse(time.RFC3339, rawDate)
	if err == nil {
		return parsedDate.Format("2006-01-02")
	}

	return rawDate
}
