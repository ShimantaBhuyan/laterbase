package fetcher

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// Crawl4AIFetcher fetches URLs using the Crawl4AI service.
type Crawl4AIFetcher struct {
	client *resty.Client
	url    string
}

// NewCrawl4AIFetcher creates a new Crawl4AIFetcher.
func NewCrawl4AIFetcher(crawl4aiURL string) *Crawl4AIFetcher {
	return &Crawl4AIFetcher{
		client: resty.New(),
		url:    crawl4aiURL,
	}
}

// Crawl4AIRequest is the request body for the Crawl4AI API.
type Crawl4AIRequest struct {
	URL               string `json:"url"`
	WordCountThreshold int   `json:"word_count_threshold"`
	ExtractionStrategy string `json:"extraction_strategy"`
}

// Crawl4AIResponse is the full response from Crawl4AI API.
type Crawl4AIResponse struct {
	Result Crawl4AIResult `json:"result"`
}

// Crawl4AIResult holds the extracted content from Crawl4AI.
type Crawl4AIResult struct {
	Title    string `json:"title"`
	Markdown string `json:"markdown"`
}

// Fetch implements the Fetcher interface for Crawl4AI.
func (f *Crawl4AIFetcher) Fetch(ctx context.Context, url string) (*FetchResult) {
	req := Crawl4AIRequest{
		URL:                url,
		WordCountThreshold: 10,
		ExtractionStrategy: "NoExtractionStrategy",
	}

	resp, err := f.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&Crawl4AIResponse{}).
		Post(f.url + "/crawl")

	if err != nil {
		return &FetchResult{Error: fmt.Errorf("crawl4ai request: %w", err)}
	}

	if resp.IsError() {
		return &FetchResult{Error: fmt.Errorf("crawl4ai status %d: %s", resp.StatusCode(), string(resp.Body()))}
	}

	result, ok := resp.Result().(*Crawl4AIResponse)
	if !ok || result == nil {
		return &FetchResult{Error: fmt.Errorf("crawl4ai: unexpected response format")}
	}

	return &FetchResult{
		Title:       result.Result.Title,
		ContentText: result.Result.Markdown,
	}
}
