package runner

import (
	"io"
	"os/exec"
)

type Runner interface {
	Run(name string, args []string, stdout, stderr io.Writer) error
}

type ExecRunner struct{}

func (ExecRunner) Run(name string, args []string, stdout, stderr io.Writer) error {
	command := exec.Command(name, args...)
	command.Stdout = stdout
	command.Stderr = stderr
	return command.Run()
}
