package modindex

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
}

type Fetcher interface {
	Fetch(ctx context.Context, request Request) ([]Entry, error)
}

type HTTPFetcher struct {
	BaseURL string
	Client  *http.Client
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
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	client := f.Client
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("index request failed: %s", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
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
