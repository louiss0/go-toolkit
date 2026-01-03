package cmd

import (
	"strings"

	"github.com/louiss0/cobra-cli-template/internal/cmdutil"
	"github.com/louiss0/cobra-cli-template/internal/runner"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func NewTestCmd(commandRunner runner.Runner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test [packages...]",
		Short: "Run Go tests (defaults to ./...)",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			goArgs := append([]string{"test"}, lo.Ternary(len(args) == 0, []string{"./..."}, args)...)

			cmdutil.LogInfoIfProduction("test: running go %s", strings.Join(goArgs, " "))
			if err := commandRunner.Run(cmd, "go", goArgs...); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
