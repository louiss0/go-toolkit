package build_info

import . "github.com/onsi/ginkgo/v2"

var _ = Describe("Build Info", func() {
	It("normalizes versions by removing the v prefix", func() {
		value := normalizeVersion("v1.2.3")

		assert.Equal("1.2.3", value)
	})

	It("keeps versions without a v prefix", func() {
		value := normalizeVersion("1.2.3")

		assert.Equal("1.2.3", value)
	})

	It("formats RFC3339 dates as yyyy-mm-dd", func() {
		value := normalizeBuildDate("2026-03-20T12:34:56Z")

		assert.Equal("2026-03-20", value)
	})

	It("keeps unknown when no build date was set", func() {
		value := normalizeBuildDate("unknown")

		assert.Equal("unknown", value)
	})

	It("keeps non-rfc3339 dates unchanged", func() {
		value := normalizeBuildDate("2026-03-20")

		assert.Equal("2026-03-20", value)
	})
})
