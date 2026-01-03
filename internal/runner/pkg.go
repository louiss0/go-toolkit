package runner

import (
	"os/exec"

	"github.com/spf13/cobra"
)

type Runner interface {
	Run(cmd *cobra.Command, name string, args ...string) error
}

type ExecRunner struct{}

func (ExecRunner) Run(cmd *cobra.Command, name string, args ...string) error {
	command := exec.Command(name, args...)
	command.Stdin = cmd.InOrStdin()
	command.Stdout = cmd.OutOrStdout()
	command.Stderr = cmd.ErrOrStderr()
	return command.Run()
}
