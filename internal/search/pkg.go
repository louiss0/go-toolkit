package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/louiss0/cobra-cli-template/custom_errors"
	"github.com/samber/lo"
)

const (
	defaultProvider = "github.com"
	proxyBaseURL    = "https://proxy.golang.org"
)

func ResolveModulePath(query string) string {
	if strings.HasPrefix(query, defaultProvider+"/") {
		return query
	}

	return defaultProvider + "/" + query
}

func FetchModuleVersions(ctx context.Context, modulePath string) ([]string, error) {
	endpoint := fmt.Sprintf("%s/%s/@v/list", proxyBaseURL, modulePath)
	client := resty.New()

	res, err := client.R().
		SetContext(ctx).
		Get(endpoint)
	if err != nil {
		return nil, custom_errors.CreateInvalidArgumentErrorWithMessage(
			fmt.Sprintf("failed to fetch module versions for %s: %v", modulePath, err),
		)
	}

	return lo.Map(strings.Split(res.String(), "\n"), func(version string, _ int) string {
		return fmt.Sprintf("%s/%s/%s", proxyBaseURL, modulePath, version)
	}), nil
}
