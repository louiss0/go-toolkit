package cmd_test

import (
	"github.com/louiss0/go-toolkit/cmd"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var Search = Describe("search command", func() {
	assert := assert.New(GinkgoT())

	It("accepts a scope and package query", func() {
		searchCmd := cmd.NewSearchCmd()

		err := searchCmd.Args(searchCmd, []string{"acme/tool"})

		assert.NoError(err)
	})

	It("rejects missing query input", func() {
		searchCmd := cmd.NewSearchCmd()

		err := searchCmd.Args(searchCmd, []string{})

		assert.Error(err)
		assert.Contains(err.Error(), "accepts 1 arg(s)")
	})

	It("rejects queries without a scope", func() {
		searchCmd := cmd.NewSearchCmd()

		err := searchCmd.Args(searchCmd, []string{"tool"})

		assert.Error(err)
		assert.Contains(err.Error(), "scope/package")
	})

	It("accepts queries with extra path segments", func() {
		searchCmd := cmd.NewSearchCmd()

		err := searchCmd.Args(searchCmd, []string{"github.com/acme/tool"})

		assert.NoError(err)
	})

	It("accepts prefixed queries with multiple segments", func() {
		searchCmd := cmd.NewSearchCmd()

		err := searchCmd.Args(searchCmd, []string{"github.com/acme/tool/extra"})

		assert.NoError(err)
	})

	It("accepts versioned short package queries", func() {
		searchCmd := cmd.NewSearchCmd()

		err := searchCmd.Args(searchCmd, []string{"onsi/ginkgo/v2"})

		assert.NoError(err)
	})
})
