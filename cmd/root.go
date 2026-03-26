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
	"context"

	"github.com/carapace-sh/carapace"
	"github.com/charmbracelet/fang"
	"github.com/kaptinlin/gozod"
	"github.com/louiss0/g-tools/mode"
	"github.com/louiss0/go-toolkit/build_info"
	"github.com/louiss0/go-toolkit/custom_errors"
	"github.com/louiss0/go-toolkit/internal/modindex/config"
	"github.com/louiss0/go-toolkit/internal/prompt"
	"github.com/louiss0/go-toolkit/internal/runner"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type RootOptions struct {
	Runner       runner.Runner
	PromptRunner prompt.Runner
	ConfigPath   string `gozod:"regex=^$|^\\S+$"`
}

var rootOptionsSchema = gozod.FromStruct[RootOptions]().
	Refine(func(options RootOptions) bool {

		return lo.EveryBy(
			[]any{options.Runner, options.PromptRunner},
			func(value any) bool {
				return value != nil
			})

	})

func NewRootCmd() *cobra.Command {
	return NewRootCmdWithOptions(RootOptions{
		Runner:       runner.ExecRunner{},
		PromptRunner: prompt.NewRunner(mode.NewModeOperator()),
	})
}

func NewRootCmdWithOptions(options RootOptions) *cobra.Command {
	if _, err := rootOptionsSchema.Parse(options); err != nil {
		panic(custom_errors.FromZod(err, custom_errors.ZodTheme{
			Subject:     "go scaffolding setup",
			RootMessage: "command wiring requires a runner and prompt runner",
			FieldMessages: map[string]string{
				"ConfigPath": "config path must not contain spaces",
			},
		}))
	}

	commandRunner := options.Runner
	promptRunner := options.PromptRunner

	configPath := config.ResolveConfigPath(options.ConfigPath)

	cmd := &cobra.Command{
		Use:   "go-toolkit",
		Short: "Go package delegation and scaffolding",
		Long: `Go Toolkit is a helper for delegating Go module workflows.
It shortens common tasks like init, remove, and scaffold.`,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	cmd.AddGroup(
		&cobra.Group{ID: "setup", Title: "Setup Commands"},
		&cobra.Group{ID: "local-packages", Title: "Local Package Commands"},
		&cobra.Group{ID: "global-packages", Title: "Global Package Commands"},
		&cobra.Group{ID: "project", Title: "Project Commands"},
	)

	cmd.PersistentFlags().StringVar(&configPath, "config", configPath, "config file path")

	initCmd := NewInitCmd(commandRunner, promptRunner, &configPath)
	addCmd := NewAddCmd(commandRunner, promptRunner, &configPath)
	removeCmd := NewRemoveCmd(commandRunner, &configPath)
	scaffoldCmd := NewScaffoldCmd(commandRunner, &configPath)
	testCmd := NewTestCmd(commandRunner)
	configCmd := NewConfigCmd(commandRunner, &configPath, promptRunner)
	searchCmd := NewSearchCmd()
	installCmd := NewInstallCmd(commandRunner, promptRunner, &configPath)
	uninstallCmd := NewUninstallCmd(commandRunner, promptRunner, &configPath)
	installGlobalsCmd := NewInstallGlobalsCmd(commandRunner, &configPath)
	initCmd.GroupID = "setup"
	configCmd.GroupID = "setup"
	addCmd.GroupID = "local-packages"
	removeCmd.GroupID = "local-packages"
	installCmd.GroupID = "global-packages"
	uninstallCmd.GroupID = "global-packages"
	installGlobalsCmd.GroupID = "global-packages"
	scaffoldCmd.GroupID = "project"
	testCmd.GroupID = "project"
	searchCmd.GroupID = "project"

	cmd.AddCommand(
		initCmd,
		addCmd,
		removeCmd,
		scaffoldCmd,
		testCmd,
		configCmd,
		searchCmd,
		installCmd,
		uninstallCmd,
		installGlobalsCmd,
	)

	configureCompletions(cmd, scaffoldCmd, configCmd)

	return cmd
}

var rootCmd = NewRootCmd()

func Execute() error {
	return fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithVersion(build_info.Version()),
		fang.WithoutCompletions(),
	)
}

func configureCompletions(root *cobra.Command, scaffoldCmd *cobra.Command, configCmd *cobra.Command) {
	rootCarapace := carapace.Gen(root)
	rootCarapace.FlagCompletion(carapace.ActionMap{
		"config": carapace.ActionFiles(".toml"),
	})

	carapace.Gen(scaffoldCmd).FlagCompletion(carapace.ActionMap{
		"folder": carapace.ActionDirectories(),
		"site":   carapace.ActionValues(config.KnownSites()...),
	})

	configCommands := configCmd.Commands()
	if len(configCommands) == 0 {
		return
	}

	for _, command := range configCommands {
		switch command.Name() {
		case "set-site":
			carapace.Gen(command).PositionalCompletion(
				carapace.ActionValues(config.KnownSites()...),
			)
		case "provider":
			configureProviderCompletions(command)
		}
	}
}

func configureProviderCompletions(providerCmd *cobra.Command) {
	for _, command := range providerCmd.Commands() {
		if command.Name() != "add" {
			continue
		}

		carapace.Gen(command).FlagCompletion(carapace.ActionMap{
			"path": carapace.ActionFiles(),
		})
	}
}
