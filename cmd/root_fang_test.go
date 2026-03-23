package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/charmbracelet/fang"
	"github.com/louiss0/go-toolkit/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestExecuteRootCommandWithFang(t *testing.T) {
	assert := assert.New(t)

	rootCmd := NewRootCmdWithOptions(RootOptions{
		Runner:       &testhelpers.RunnerMock{},
		PromptRunner: testhelpers.NewPromptRunnerMock(),
	})

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"--help"})

	err := fang.Execute(context.Background(), rootCmd, fang.WithoutCompletions())

	output := stdout.String() + stderr.String()

	assert.NoError(err)
	assert.True(rootCmd.CompletionOptions.DisableDefaultCmd)
	assert.Contains(output, "Go Toolkit is a helper for delegating Go module workflows.")
	assert.Contains(output, "config")
	assert.Contains(output, "scaffold")
	assert.NotContains(output, "\n  completion")

	completionCmd, _, completionErr := rootCmd.Find([]string{"completion"})
	assert.Error(completionErr)
	assert.NotNil(completionCmd)
	assert.Equal("go-toolkit", completionCmd.Name())

	carapaceCmd, _, carapaceErr := rootCmd.Find([]string{"_carapace"})
	assert.NoError(carapaceErr)
	assert.NotNil(carapaceCmd)
	assert.Equal("_carapace", carapaceCmd.Name())
	assert.True(carapaceCmd.Hidden)
}

func TestNewRootCmdWithOptionsRegistersHiddenCarapaceCommand(t *testing.T) {
	assert := assert.New(t)

	rootCmd := NewRootCmdWithOptions(RootOptions{
		Runner:       &testhelpers.RunnerMock{},
		PromptRunner: testhelpers.NewPromptRunnerMock(),
	})

	carapaceCmd, _, err := rootCmd.Find([]string{"_carapace"})

	assert.NoError(err)
	assert.NotNil(carapaceCmd)
	assert.True(carapaceCmd.Hidden)
}
