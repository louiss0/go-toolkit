package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kaptinlin/gozod"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

const DefaultSite = "github.com"

var knownSites = map[string]struct{}{
	"github.com":    {},
	"gitlab.com":    {},
	"bitbucket.org": {},
}

type Values struct {
	User      string           `mapstructure:"user" toml:"user" gozod:"regex=^\\S*$"`
	Site      string           `mapstructure:"site" toml:"site" gozod:"regex=^$|^[^\\s.][^\\s]*\\.[^\\s]*[^\\s.]$"`
	Scaffold  ScaffoldConfig   `mapstructure:"scaffold" toml:"scaffold"`
	Providers []ProviderConfig `mapstructure:"providers" toml:"providers"`
}

type ProviderConfig struct {
	Name string `mapstructure:"name" toml:"name" gozod:"required,min=1"`
	Path string `mapstructure:"path" toml:"path" gozod:"required,min=1"`
}

type ScaffoldConfig struct {
	WriteTests bool `mapstructure:"write_tests" toml:"write_tests"`
}

var valuesSchema = gozod.FromStruct[Values]()
var siteSchema = gozod.String().Regex(regexp.MustCompile(`^[^\s.][^\s]*\.[^\s]*[^\s.]$`))

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

	if err := validateValues(values); err != nil {
		return Values{}, err
	}

	return values, nil
}

func Save(path string, values Values) error {
	if path == "" {
		return errors.New("config path is required")
	}

	if err := validateValues(values); err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	configFile := viper.New()
	configFile.SetConfigType("toml")
	configFile.Set("user", values.User)
	configFile.Set("site", values.Site)
	configFile.Set("scaffold", values.Scaffold)
	if len(values.Providers) > 0 {
		configFile.Set("providers", values.Providers)
	}

	return configFile.WriteConfigAs(path)
}

func validateValues(values Values) error {
	if _, err := valuesSchema.Parse(values); err != nil {
		return fmt.Errorf("invalid config values: %w", err)
	}

	return nil
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
	return lo.MapToSlice(knownSites, func(key string, value struct{}) string {
		return key
	})
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
