package config

import (
	"os"
	"path/filepath"
)

func ResolveConfigPath(configPath string) string {
	if configPath != "" {
		return configPath
	}

	localPath := LocalConfigPath()
	if localPath != "" {
		return localPath
	}

	defaultPath, err := DefaultPath()
	if err != nil {
		return ""
	}

	return defaultPath
}

func LocalConfigPath() string {
	workingDir, err := os.Getwd()
	if err != nil {
		return ""
	}

	path := filepath.Join(workingDir, "gtk-config.toml")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	return ""
}
