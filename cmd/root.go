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

	"github.com/kaptinlin/gozod"
	"github.com/louiss0/cobra-cli-template/internal/config"
	"github.com/louiss0/cobra-cli-template/internal/prompt"
	"github.com/louiss0/cobra-cli-template/internal/runner"
	"github.com/louiss0/g-tools/mode"
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

		return lo.EveryBy([]any{options.Runner, options.PromptRunner}, func(value any) bool {
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

	if options.Runner == nil {
		options.Runner = runner.ExecRunner{}
	}
	if options.PromptRunner == nil {
		options.PromptRunner = prompt.NewRunner(mode.NewModeOperator())
	}

	rootOptionsSchema.MustParse(options)

	commandRunner := options.Runner
	promptRunner := options.PromptRunner

	configPath := config.ResolveConfigPath(options.ConfigPath)

	cmd := &cobra.Command{
		Use:   "go-toolkit",
		Short: "Go package delegation and scaffolding",
		Long: `Go Toolkit is a helper for delegating Go module workflows.
It shortens common tasks like init, remove, and scaffold.`,
	}

	cmd.PersistentFlags().StringVar(&configPath, "config", configPath, "config file path")

	cmd.AddCommand(NewInitCmd(commandRunner, promptRunner, &configPath))
	cmd.AddCommand(NewAddCmd(commandRunner, &configPath))
	cmd.AddCommand(NewRemoveCmd(commandRunner, &configPath))
	cmd.AddCommand(NewScaffoldCmd(commandRunner, &configPath))
	cmd.AddCommand(NewTestCmd(commandRunner))
	cmd.AddCommand(NewConfigCmd(&configPath, promptRunner))
	cmd.AddCommand(NewSearchCmd())

	return cmd
}

var rootCmd = NewRootCmd()

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
