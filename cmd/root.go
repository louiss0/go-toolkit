/*
Copyright © 2025 Shelton Louis

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"os"
	"path/filepath"

	"github.com/louiss0/cobra-cli-template/internal/config"
	"github.com/louiss0/cobra-cli-template/internal/modindex"
	"github.com/louiss0/cobra-cli-template/internal/runner"
	"github.com/spf13/cobra"
)

type RootOptions struct {
	Runner       runner.Runner
	ConfigPath   string
	IndexFetcher modindex.Fetcher
}

func NewRootCmd() *cobra.Command {
	return NewRootCmdWithOptions(RootOptions{})
}

func NewRootCmdWithOptions(options RootOptions) *cobra.Command {
	commandRunner := options.Runner
	if commandRunner == nil {
		commandRunner = runner.ExecRunner{}
	}

	indexFetcher := options.IndexFetcher
	if indexFetcher == nil {
		indexFetcher = modindex.HTTPFetcher{}
	}

	configPath := resolveConfigPath(options.ConfigPath)

	cmd := &cobra.Command{
		Use:   "go-toolkit",
		Short: "Go package delegation and scaffolding",
		Long: `Go Toolkit is a helper for delegating Go module workflows.
It shortens common tasks like init, remove, and scaffold.`,
	}

	cmd.PersistentFlags().StringVar(&configPath, "config", configPath, "config file path")

	cmd.AddCommand(NewInitCmd(commandRunner, &configPath))
	cmd.AddCommand(NewAddCmd(commandRunner, &configPath))
	cmd.AddCommand(NewRemoveCmd(commandRunner, &configPath))
	cmd.AddCommand(NewScaffoldCmd(commandRunner, &configPath))
	cmd.AddCommand(NewTestCmd(commandRunner))
	cmd.AddCommand(NewConfigCmd(&configPath))
	cmd.AddCommand(NewSearchCmd(indexFetcher, &configPath))

	return cmd
}

func resolveConfigPath(configPath string) string {
	if configPath != "" {
		return configPath
	}

	localPath := localConfigPath()
	if localPath != "" {
		return localPath
	}

	defaultPath, err := config.DefaultPath()
	if err != nil {
		return ""
	}

	return defaultPath
}

func localConfigPath() string {
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

var rootCmd = NewRootCmd()

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
