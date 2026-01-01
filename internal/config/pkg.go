package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const DefaultSite = "github.com"

var knownSites = map[string]struct{}{
	"github.com":    {},
	"gitlab.com":    {},
	"bitbucket.org": {},
}

type Values struct {
	User      string           `mapstructure:"user" toml:"user"`
	Site      string           `mapstructure:"site" toml:"site"`
	Providers []ProviderConfig `mapstructure:"providers" toml:"providers"`
}

type ProviderConfig struct {
	Name string `mapstructure:"name" toml:"name"`
	Path string `mapstructure:"path" toml:"path"`
}

func DefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "go-toolkit", "gtk-config.toml"), nil
}

func Load(path string) (Values, error) {
	if path == "" {
		return Values{}, errors.New("config path is required")
	}

	configFile := viper.New()
	configFile.SetConfigFile(path)
	configFile.SetConfigType("toml")

	if err := configFile.ReadInConfig(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Values{}, nil
		}
		return Values{}, err
	}

	var values Values
	if err := configFile.Unmarshal(&values); err != nil {
		return Values{}, err
	}

	return values, nil
}

func Save(path string, values Values) error {
	if path == "" {
		return errors.New("config path is required")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	configFile := viper.New()
	configFile.SetConfigType("toml")
	configFile.Set("user", values.User)
	configFile.Set("site", values.Site)
	if len(values.Providers) > 0 {
		configFile.Set("providers", values.Providers)
	}

	return configFile.WriteConfigAs(path)
}

func ResolveSite(flagSite string, values Values) string {
	if flagSite != "" {
		return flagSite
	}

	if values.Site != "" {
		return values.Site
	}

	return DefaultSite
}

func IsKnownSite(site string) bool {
	_, ok := knownSites[site]
	return ok
}

func KnownSites() []string {
	return []string{"github.com", "gitlab.com", "bitbucket.org"}
}

func IsValidSite(site string) bool {
	trimmed := strings.TrimSpace(site)
	if trimmed == "" {
		return false
	}

	if strings.Contains(trimmed, " ") {
		return false
	}

	return strings.Contains(trimmed, ".") && !strings.HasPrefix(trimmed, ".") && !strings.HasSuffix(trimmed, ".")
}
