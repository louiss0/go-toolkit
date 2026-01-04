package cmd

import (
	"github.com/charmbracelet/huh"
	. "github.com/onsi/ginkgo/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Config init prompt", func() {
	assert := assert.New(GinkgoT())

	It("returns aborted when username prompt is canceled", func() {
		mock := newPromptMock(promptStep{
			kind: promptStepInput,
			err:  huh.ErrUserAborted,
		})

		_, err := promptConfigInitInputs(&cobra.Command{}, mock)

		assert.ErrorIs(err, huh.ErrUserAborted)
	})

	It("returns username when provider prompt is canceled", func() {
		mock := newPromptMock(
			promptStep{kind: promptStepInput, value: "lou"},
			promptStep{kind: promptStepSelect, err: huh.ErrUserAborted},
		)

		values, err := promptConfigInitInputs(&cobra.Command{}, mock)

		assert.NoError(err)
		assert.Equal("lou", values.UserName)
		assert.Equal("", values.ProviderSite)
	})

	It("supports provider skip remaining", func() {
		mock := newPromptMock(
			promptStep{kind: promptStepInput, value: "lou"},
			promptStep{kind: promptStepSelect, value: providerSkipRemaining},
		)

		values, err := promptConfigInitInputs(&cobra.Command{}, mock)

		assert.NoError(err)
		assert.Equal("lou", values.UserName)
		assert.Equal("", values.ProviderSite)
	})

	It("accepts provider skip", func() {
		mock := newPromptMock(
			promptStep{kind: promptStepInput, value: "lou"},
			promptStep{kind: promptStepSelect, value: providerSkip},
		)

		values, err := promptConfigInitInputs(&cobra.Command{}, mock)

		assert.NoError(err)
		assert.Equal("lou", values.UserName)
		assert.Equal("", values.ProviderSite)
	})

	It("accepts custom provider input", func() {
		mock := newPromptMock(
			promptStep{kind: promptStepInput, value: "lou"},
			promptStep{kind: promptStepSelect, value: providerCustom},
			promptStep{kind: promptStepInput, value: "gitlab.com"},
		)

		values, err := promptConfigInitInputs(&cobra.Command{}, mock)

		assert.NoError(err)
		assert.Equal("lou", values.UserName)
		assert.Equal("gitlab.com", values.ProviderSite)
	})
})
