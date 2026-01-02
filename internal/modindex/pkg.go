package modindex

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const DefaultBaseURL = "https://index.golang.org"

type Entry struct {
	Path      string `json:"Path"`
	Version   string `json:"Version"`
	Timestamp string `json:"Timestamp"`
}

type Request struct {
	Since      string
	Limit      int
	IncludeAll bool
	MaxPages   int
	MaxResults int
}

type Fetcher interface {
	Fetch(ctx context.Context, request Request) ([]Entry, error)
}

type HTTPFetcher struct {
	BaseURL string
	Client  *resty.Client
}

func (f HTTPFetcher) Fetch(ctx context.Context, request Request) ([]Entry, error) {
	baseURL := f.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	query := url.Values{}
	if request.Since != "" {
		query.Set("since", request.Since)
	}
	if request.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", request.Limit))
	}
	if request.IncludeAll {
		query.Set("include", "all")
	}

	endpoint := baseURL + "/index"

	client := f.Client
	if client == nil {
		client = resty.New().
			SetRetryCount(3).
			SetRetryWaitTime(200 * time.Millisecond).
			SetRetryMaxWaitTime(2 * time.Second).
			AddRetryCondition(func(resp *resty.Response, err error) bool {
				if err != nil {
					return true
				}
				return resp != nil && resp.StatusCode() >= http.StatusInternalServerError
			})
	}

	resp, err := client.R().SetContext(ctx).SetQueryParamsFromValues(query).Get(endpoint)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("index request failed: %s", resp.Status())
	}

	scanner := bufio.NewScanner(bytes.NewReader(resp.Body()))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	entries := []Entry{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry Entry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func FilterEntries(entries []Entry, query string, site string, useSiteFilter bool) []Entry {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}

	filtered := []Entry{}
	for _, entry := range entries {
		path := strings.ToLower(entry.Path)
		if useSiteFilter && site != "" {
			if !strings.HasPrefix(path, strings.ToLower(site)+"/") {
				continue
			}
		}

		if !strings.Contains(path, query) {
			continue
		}

		filtered = append(filtered, entry)
	}

	return filtered
}

func DropLeadingDuplicate(entries []Entry, previous Entry) []Entry {
	if len(entries) == 0 {
		return entries
	}

	if entries[0] == previous {
		return entries[1:]
	}

	return entries
}

func FetchAll(ctx context.Context, fetcher Fetcher, request Request) ([]Entry, error) {
	if fetcher == nil {
		return nil, fmt.Errorf("fetcher is required")
	}

	if request.MaxPages < 0 || request.MaxResults < 0 {
		return nil, fmt.Errorf("max pages and max results must be non-negative")
	}

	if request.Limit <= 0 {
		entries, err := fetcher.Fetch(ctx, request)
		if err != nil {
			return nil, err
		}

		if request.MaxResults > 0 && len(entries) > request.MaxResults {
			return entries[:request.MaxResults], nil
		}

		return entries, nil
	}

	pageRequest := request
	pageSince := request.Since

	var previous *Entry
	entries := []Entry{}
	pagesFetched := 0

	for {
		if request.MaxPages > 0 && pagesFetched >= request.MaxPages {
			break
		}

		pageRequest.Since = pageSince
		pageEntries, err := fetcher.Fetch(ctx, pageRequest)
		if err != nil {
			return nil, err
		}
		pagesFetched++

		if len(pageEntries) == 0 {
			break
		}

		filteredEntries := pageEntries
		if previous != nil {
			filteredEntries = DropLeadingDuplicate(pageEntries, *previous)
		}

		if len(filteredEntries) > 0 {
			entries = append(entries, filteredEntries...)
		} else if previous != nil && len(pageEntries) == 1 && pageEntries[0] == *previous {
			break
		}

		if request.MaxResults > 0 && len(entries) >= request.MaxResults {
			return entries[:request.MaxResults], nil
		}

		lastEntry := pageEntries[len(pageEntries)-1]
		previous = &lastEntry
		pageSince = lastEntry.Timestamp

		if len(pageEntries) < pageRequest.Limit {
			break
		}
	}

	return entries, nil
}
