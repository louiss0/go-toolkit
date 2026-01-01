package modindex_test

import (
	"github.com/louiss0/cobra-cli-template/internal/modindex"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("FilterEntries", func() {
	assert := assert.New(GinkgoT())

	DescribeTable("filters module entries",
		func(query, site string, useSiteFilter bool, expectedPath string) {
			entries := []modindex.Entry{
				{Path: "github.com/acme/tool"},
				{Path: "gitlab.com/acme/tool"},
			}

			filtered := modindex.FilterEntries(entries, query, site, useSiteFilter)

			assert.Len(filtered, 1)
			assert.Equal(expectedPath, filtered[0].Path)
		},
		Entry("filters by site when the query is short", "tool", "github.com", true, "github.com/acme/tool"),
		Entry("does not filter by site when using a full domain query", "gitlab.com", "github.com", false, "gitlab.com/acme/tool"),
	)
})
