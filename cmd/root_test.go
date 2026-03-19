package cmd_test

import (
	"github.com/louiss0/go-toolkit/cmd"
	"github.com/louiss0/go-toolkit/internal/testhelpers"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var RootOptions = Describe("root options", func() {
	assert := assert.New(GinkgoT())

	It("panics when the config path is whitespace", func() {
		assert.Panics(func() {
			_ = cmd.NewRootCmdWithOptions(cmd.RootOptions{
				Runner:       &testhelpers.RunnerMock{},
				PromptRunner: testhelpers.NewPromptRunnerMock(),
				ConfigPath:   "   ",
			})
		})
	})
})
