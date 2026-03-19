package cmd_test

import (
	"github.com/louiss0/go-toolkit/cmd"
	"github.com/louiss0/go-toolkit/internal/testhelpers"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var Test = Describe("test command", func() {
	assert := assert.New(GinkgoT())

	It("runs go test for all packages by default", func() {
		runner := &testhelpers.RunnerMock{}
		runner.On("Run", mock.Anything, "go", []string{"test", "./..."}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
		})

		_, err := testhelpers.ExecuteCmd(rootCmd, "test")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())
	})
})
