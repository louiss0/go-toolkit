package cmd

import (
	"github.com/charmbracelet/huh"
	. "github.com/onsi/ginkgo/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Init prompt", func() {
	assert := assert.New(GinkgoT())

	It("returns aborted when the module prompt is canceled", func() {
		mock := newPromptMock(promptStep{
			kind: promptStepInput,
			err:  huh.ErrUserAborted,
		})

		_, err := promptInitInputs(&cobra.Command{}, mock)

		assert.ErrorIs(err, huh.ErrUserAborted)
	})

	It("returns module name when canceled after the first step", func() {
		mock := newPromptMock(
			promptStep{kind: promptStepInput, value: "toolkit"},
			promptStep{kind: promptStepInput, err: huh.ErrUserAborted},
		)

		values, err := promptInitInputs(&cobra.Command{}, mock)

		assert.NoError(err)
		assert.Equal("toolkit", values.ModuleName)
		assert.True(values.Used)
	})

	It("supports provider skip remaining", func() {
		mock := newPromptMock(
			promptStep{kind: promptStepInput, value: "toolkit"},
			promptStep{kind: promptStepInput, value: ""},
			promptStep{kind: promptStepSelect, value: providerSkipRemaining},
		)

		values, err := promptInitInputs(&cobra.Command{}, mock)

		assert.NoError(err)
		assert.Equal("toolkit", values.ModuleName)
		assert.Equal("", values.ProviderSite)
	})

	It("handles custom provider cancellation", func() {
		mock := newPromptMock(
			promptStep{kind: promptStepInput, value: "toolkit"},
			promptStep{kind: promptStepInput, value: "lou"},
			promptStep{kind: promptStepSelect, value: providerCustom},
			promptStep{kind: promptStepInput, err: huh.ErrUserAborted},
		)

		values, err := promptInitInputs(&cobra.Command{}, mock)

		assert.NoError(err)
		assert.Equal("toolkit", values.ModuleName)
		assert.Equal("lou", values.UserName)
		assert.Equal("", values.ProviderSite)
	})

	It("accepts skip values for later steps", func() {
		mock := newPromptMock(
			promptStep{kind: promptStepInput, value: "toolkit"},
			promptStep{kind: promptStepInput, value: ""},
			promptStep{kind: promptStepSelect, value: providerSkip},
			promptStep{kind: promptStepSelect, value: projectTypeLibrary},
			promptStep{kind: promptStepSelect, value: testChoiceNo},
			promptStep{kind: promptStepSelect, value: gitChoiceNo},
			promptStep{kind: promptStepInput, value: "github.com/spf13/cobra"},
		)

		values, err := promptInitInputs(&cobra.Command{}, mock)

		assert.NoError(err)
		assert.Equal(projectTypeLibrary, values.ProjectType)
		assert.Equal(testChoiceNo, values.TestDrivenChoice)
		assert.Equal(gitChoiceNo, values.GitChoice)
		assert.Equal([]string{"github.com/spf13/cobra"}, values.Packages)
	})

	It("stops after test skip remaining", func() {
		mock := newPromptMock(
			promptStep{kind: promptStepInput, value: "toolkit"},
			promptStep{kind: promptStepInput, value: "lou"},
			promptStep{kind: promptStepSelect, value: "github.com"},
			promptStep{kind: promptStepSelect, value: projectTypeApp},
			promptStep{kind: promptStepSelect, value: testChoiceSkipRemaining},
		)

		values, err := promptInitInputs(&cobra.Command{}, mock)

		assert.NoError(err)
		assert.Equal(testChoiceSkipRemaining, values.TestDrivenChoice)
		assert.Equal("", values.GitChoice)
		assert.Nil(values.Packages)
	})

	It("stops after git skip remaining", func() {
		mock := newPromptMock(
			promptStep{kind: promptStepInput, value: "toolkit"},
			promptStep{kind: promptStepInput, value: "lou"},
			promptStep{kind: promptStepSelect, value: "github.com"},
			promptStep{kind: promptStepSelect, value: projectTypeApp},
			promptStep{kind: promptStepSelect, value: testChoiceYes},
			promptStep{kind: promptStepSelect, value: gitChoiceSkipRemaining},
		)

		values, err := promptInitInputs(&cobra.Command{}, mock)

		assert.NoError(err)
		assert.Equal(gitChoiceSkipRemaining, values.GitChoice)
	})

	It("allows canceling packages input", func() {
		mock := newPromptMock(
			promptStep{kind: promptStepInput, value: "toolkit"},
			promptStep{kind: promptStepInput, value: "lou"},
			promptStep{kind: promptStepSelect, value: "github.com"},
			promptStep{kind: promptStepSelect, value: projectTypeApp},
			promptStep{kind: promptStepSelect, value: testChoiceYes},
			promptStep{kind: promptStepSelect, value: gitChoiceYes},
			promptStep{kind: promptStepInput, err: huh.ErrUserAborted},
		)

		values, err := promptInitInputs(&cobra.Command{}, mock)

		assert.NoError(err)
		assert.Nil(values.Packages)
	})
})
