package cmd

import (
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

			return commandRunner.Run("go", goArgs, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}

	return cmd
}
