package recsys

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	trimmed := strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL: trimmed,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) GetFeed(memberID int64) ([]string, error) {
	endpoint := fmt.Sprintf("%s/feeds?member_id=%d", c.baseURL, memberID)
	return c.fetchArticleIDs(endpoint)
}

func (c *Client) Search(query string) ([]string, error) {
	endpoint := fmt.Sprintf("%s/search?query=%s", c.baseURL, url.QueryEscape(query))
	return c.fetchArticleIDs(endpoint)
}

func (c *Client) fetchArticleIDs(endpoint string) ([]string, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("recsys request failed with status %d", resp.StatusCode)
	}

	var payload struct {
		Articles []string `json:"articles"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return payload.Articles, nil
}
