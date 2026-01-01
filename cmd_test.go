package main

import (
	"bytes"
	"fmt"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/louiss0/cobra-cli-template/cmd"
	. "github.com/onsi/ginkgo/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// This function executes a cobra command with the given arguments and returns the output and error.
// It sets the output and error buffers for the command, sets the arguments, and executes the command.
// If there is an error, it returns an error with the error message from the error buffer.
// If there is no error, it returns the output from the output buffer.
// It's used to test the cobra commands.
// When you use this function, make sure to pass the root command and any arguments you want to test.
// The first argument after the rootCmd is any sub command or flag you want to test.
func executeCmd(cmd *cobra.Command, args ...string) (string, error) {

	buf := new(bytes.Buffer)
	errBuff := new(bytes.Buffer)

	cmd.SetOut(buf)
	cmd.SetErr(errBuff)
	cmd.SetArgs(args)

	err := cmd.Execute()

	if errBuff.Len() > 0 {
		return "", fmt.Errorf("command failed: %s", errBuff.String())
	}

	return buf.String(), err
}

var rootCmd = cmd.NewRootCmd()

var _ = Describe("Cmd", func() {

	assert := assert.New(GinkgoT())

	It("should be able to run", func() {

		// The "" needs to be passed as an argument to the executeCmd function with rootCmd
		// If not then there will be an error
		// This is because gingko will pass a flag to the command to indicate that it is running in a test environment
		_, err := executeCmd(rootCmd, "")

		assert.NoError(err)

	})

	It("should give me a random number", func() {

		number := gofakeit.Number(1, 100)

		assert.True(number >= 1 && number <= 100)

		assert.Equal(number, gofakeit.Number(1, 100))
	})

})
