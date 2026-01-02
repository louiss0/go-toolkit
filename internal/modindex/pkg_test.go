package modindex_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/go-resty/resty/v2"
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

var _ = Describe("DropLeadingDuplicate", func() {
	assert := assert.New(GinkgoT())

	It("drops the first entry when it matches the previous entry", func() {
		previous := modindex.Entry{Path: "github.com/acme/tool", Version: "v1.2.3", Timestamp: "2024-01-01T00:00:00Z"}
		entries := []modindex.Entry{
			previous,
			{Path: "github.com/acme/next", Version: "v1.0.0", Timestamp: "2024-01-01T01:00:00Z"},
		}

		filtered := modindex.DropLeadingDuplicate(entries, previous)

		assert.Len(filtered, 1)
		assert.Equal("github.com/acme/next", filtered[0].Path)
	})

	It("keeps entries when the first entry does not match the previous entry", func() {
		previous := modindex.Entry{Path: "github.com/acme/tool", Version: "v1.2.3", Timestamp: "2024-01-01T00:00:00Z"}
		entries := []modindex.Entry{
			{Path: "github.com/acme/next", Version: "v1.0.0", Timestamp: "2024-01-01T01:00:00Z"},
		}

		filtered := modindex.DropLeadingDuplicate(entries, previous)

		assert.Len(filtered, 1)
		assert.Equal("github.com/acme/next", filtered[0].Path)
	})
})

var _ = Describe("HTTPFetcher", func() {
	assert := assert.New(GinkgoT())

	It("requests the index with query params and parses entries", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal("/index", r.URL.Path)
			assert.Equal("2019-04-10T20:30:02.04035Z", r.URL.Query().Get("since"))
			assert.Equal("2", r.URL.Query().Get("limit"))
			assert.Equal("all", r.URL.Query().Get("include"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{\"Path\":\"github.com/acme/tool\",\"Version\":\"v1.0.0\",\"Timestamp\":\"2019-04-10T20:30:02.04035Z\"}\n"))
			_, _ = w.Write([]byte("{\"Path\":\"github.com/acme/next\",\"Version\":\"v1.1.0\",\"Timestamp\":\"2019-04-10T20:40:02.04035Z\"}\n"))
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL).SetTransport(server.Client().Transport)
		fetcher := modindex.HTTPFetcher{
			BaseURL: server.URL,
			Client:  client,
		}

		entries, err := fetcher.Fetch(context.Background(), modindex.Request{
			Since:      "2019-04-10T20:30:02.04035Z",
			Limit:      2,
			IncludeAll: true,
		})

		assert.NoError(err)
		assert.Len(entries, 2)
		assert.Equal("github.com/acme/tool", entries[0].Path)
		assert.Equal("github.com/acme/next", entries[1].Path)
	})

	It("returns an error for non-200 responses", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL).SetTransport(server.Client().Transport)
		fetcher := modindex.HTTPFetcher{
			BaseURL: server.URL,
			Client:  client,
		}

		_, err := fetcher.Fetch(context.Background(), modindex.Request{})

		assert.Error(err)
		assert.Contains(err.Error(), "index request failed")
	})
})

type fetcherStub struct {
	responses [][]modindex.Entry
	calls     []modindex.Request
}

func (f *fetcherStub) Fetch(ctx context.Context, request modindex.Request) ([]modindex.Entry, error) {
	f.calls = append(f.calls, request)
	if len(f.responses) == 0 {
		return nil, nil
	}

	response := f.responses[0]
	f.responses = f.responses[1:]
	return response, nil
}

var _ = Describe("FetchAll", func() {
	assert := assert.New(GinkgoT())

	It("paginates and de-duplicates inclusive entries", func() {
		stub := &fetcherStub{
			responses: [][]modindex.Entry{
				{
					{Path: "github.com/acme/first", Version: "v1.0.0", Timestamp: "2024-01-01T00:00:00Z"},
					{Path: "github.com/acme/second", Version: "v1.1.0", Timestamp: "2024-01-01T00:10:00Z"},
				},
				{
					{Path: "github.com/acme/second", Version: "v1.1.0", Timestamp: "2024-01-01T00:10:00Z"},
					{Path: "github.com/acme/third", Version: "v1.2.0", Timestamp: "2024-01-01T00:20:00Z"},
				},
			},
		}

		entries, err := modindex.FetchAll(context.Background(), stub, modindex.Request{
			Limit: 2,
		})

		assert.NoError(err)
		assert.Len(entries, 3)
		assert.Equal("github.com/acme/first", entries[0].Path)
		assert.Equal("github.com/acme/second", entries[1].Path)
		assert.Equal("github.com/acme/third", entries[2].Path)
		assert.Len(stub.calls, 3)
		assert.Equal("", stub.calls[0].Since)
		assert.Equal("2024-01-01T00:10:00Z", stub.calls[1].Since)
		assert.Equal("2024-01-01T00:20:00Z", stub.calls[2].Since)
	})

	It("returns a single page when no limit is set", func() {
		stub := &fetcherStub{
			responses: [][]modindex.Entry{
				{
					{Path: "github.com/acme/first", Version: "v1.0.0", Timestamp: "2024-01-01T00:00:00Z"},
					{Path: "github.com/acme/second", Version: "v1.1.0", Timestamp: "2024-01-01T00:10:00Z"},
				},
			},
		}

		entries, err := modindex.FetchAll(context.Background(), stub, modindex.Request{})

		assert.NoError(err)
		assert.Len(entries, 2)
		assert.Len(stub.calls, 1)
	})
})
