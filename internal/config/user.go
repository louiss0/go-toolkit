package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

var ErrMissingUser = errors.New("missing user")

func ResolveUser(flagUser string, values Values, site string) (string, error) {
	if flagUser != "" {
		return flagUser, nil
	}

	if values.User != "" {
		return values.User, nil
	}

	providerName := providerNameForSite(site)
	configPath := providerConfigPath(values.Providers, providerName)
	if configPath == "" && site == DefaultSite {
		defaultPath, err := defaultGitConfigPath()
		if err != nil {
			return "", err
		}
		configPath = defaultPath
	}

	if configPath == "" {
		return "", ErrMissingUser
	}

	userName, err := readGitUserName(configPath)
	if err != nil {
		return "", err
	}

	if userName == "" {
		return "", ErrMissingUser
	}

	return userName, nil
}

func providerNameForSite(site string) string {
	trimmed := strings.TrimSpace(site)
	switch trimmed {
	case "github.com":
		return "github"
	case "gitlab.com":
		return "gitlab"
	case "bitbucket.org":
		return "bitbucket"
	default:
		return trimmed
	}
}

func providerConfigPath(providers []ProviderConfig, name string) string {
	for _, provider := range providers {
		if provider.Name == name {
			return provider.Path
		}
	}

	return ""
}

func defaultGitConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".gitconfig"), nil
}

func readGitUserName(path string) (string, error) {
	configFile, err := ini.Load(path)
	if err != nil {
		return "", fmt.Errorf("read git config: %w", err)
	}

	section, err := configFile.GetSection("user")
	if err != nil {
		return "", ErrMissingUser
	}

	key, err := section.GetKey("name")
	if err != nil {
		return "", ErrMissingUser
	}

	return strings.TrimSpace(key.String()), nil
}
