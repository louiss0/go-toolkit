package modindex_test

import (
	"github.com/louiss0/cobra-cli-template/internal/modindex"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("FilterEntries", func() {
	assert := assert.New(GinkgoT())

	It("filters by site when the query is short", func() {
		entries := []modindex.Entry{
			{Path: "github.com/acme/tool"},
			{Path: "gitlab.com/acme/tool"},
		}

		filtered := modindex.FilterEntries(entries, "tool", "github.com", true)

		assert.Len(filtered, 1)
		assert.Equal("github.com/acme/tool", filtered[0].Path)
	})

	It("does not filter by site when using a full domain query", func() {
		entries := []modindex.Entry{
			{Path: "github.com/acme/tool"},
			{Path: "gitlab.com/acme/tool"},
		}

		filtered := modindex.FilterEntries(entries, "gitlab.com", "github.com", false)

		assert.Len(filtered, 1)
		assert.Equal("gitlab.com/acme/tool", filtered[0].Path)
	})
})
