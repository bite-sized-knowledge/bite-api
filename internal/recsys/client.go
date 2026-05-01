package recsys

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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

// FeedResult bundles the recsys /feeds response. FeedRequestID 는 옵션이며
// bite-web 가 user_events.feed_request_id 로 echo 하면 impression ↔ click 정확 그룹핑.
type FeedResult struct {
	Articles      []string
	FeedRequestID string
}

// GetFeed asks recsys for a personalized feed. Pass memberID > 0 for
// authenticated callers; for anonymous callers, pass 0 and provide deviceID.
// recsys requires at least one of the two identifiers.
func (c *Client) GetFeed(memberID int64, deviceID string) (FeedResult, error) {
	q := url.Values{}
	if memberID > 0 {
		q.Set("member_id", strconv.FormatInt(memberID, 10))
	}
	if deviceID != "" {
		q.Set("device_id", deviceID)
	}
	if len(q) == 0 {
		return FeedResult{}, fmt.Errorf("recsys feed: member_id or device_id required")
	}
	endpoint := fmt.Sprintf("%s/feeds?%s", c.baseURL, q.Encode())
	return c.fetchFeed(endpoint)
}

func (c *Client) fetchFeed(endpoint string) (FeedResult, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return FeedResult{}, err
	}
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return FeedResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return FeedResult{}, fmt.Errorf("recsys feed failed with status %d", resp.StatusCode)
	}
	var payload struct {
		Articles      []string `json:"articles"`
		FeedRequestID string   `json:"feed_request_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return FeedResult{}, err
	}
	return FeedResult{Articles: payload.Articles, FeedRequestID: payload.FeedRequestID}, nil
}

// PostFeedback fires a single bandit/user-vector update. Pass memberID > 0 for
// authenticated callers; for anonymous callers, pass 0 and provide deviceID.
// recsys requires at least one of the two identifiers (member_id or device_id).
// Used as fire-and-forget from event.Service.Create — failures are logged by the caller.
func (c *Client) PostFeedback(memberID int64, deviceID, articleID, eventType string) error {
	if (memberID <= 0 && deviceID == "") || articleID == "" || eventType == "" {
		return fmt.Errorf("invalid feedback args")
	}
	type payload struct {
		MemberID  *int64 `json:"member_id,omitempty"`
		DeviceID  string `json:"device_id,omitempty"`
		ArticleID string `json:"article_id"`
		EventType string `json:"event_type"`
	}
	p := payload{ArticleID: articleID, EventType: eventType}
	if memberID > 0 {
		p.MemberID = &memberID
	} else if deviceID != "" {
		p.DeviceID = deviceID
	}
	body, err := json.Marshal(p)
	if err != nil {
		return err
	}
	endpoint := fmt.Sprintf("%s/feeds/feedback", c.baseURL)
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("recsys feedback failed with status %d", resp.StatusCode)
	}
	return nil
}

type SearchRequest struct {
	Query           string
	Limit           int
	From            string
	CategoryID      *int64
	Lang            string
	BlogID          *int64
	PublishedAfter  *float64
	PublishedBefore *float64
	Mode            string
}

type SearchResult struct {
	Articles []string
	Next     string
	QueryID  string
}

func (c *Client) Search(req SearchRequest) (SearchResult, error) {
	q := url.Values{}
	q.Set("query", req.Query)
	if req.Limit > 0 {
		q.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.From != "" {
		q.Set("cursor", req.From)
	}
	if req.CategoryID != nil {
		q.Set("category_id", strconv.FormatInt(*req.CategoryID, 10))
	}
	if req.Lang != "" {
		q.Set("lang", req.Lang)
	}
	if req.BlogID != nil {
		q.Set("blog_id", strconv.FormatInt(*req.BlogID, 10))
	}
	if req.PublishedAfter != nil {
		q.Set("published_after", strconv.FormatFloat(*req.PublishedAfter, 'f', -1, 64))
	}
	if req.PublishedBefore != nil {
		q.Set("published_before", strconv.FormatFloat(*req.PublishedBefore, 'f', -1, 64))
	}
	if req.Mode != "" {
		q.Set("mode", req.Mode)
	}

	endpoint := fmt.Sprintf("%s/search?%s", c.baseURL, q.Encode())
	httpReq, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return SearchResult{}, err
	}
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	}
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return SearchResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return SearchResult{}, fmt.Errorf("recsys search failed with status %d", resp.StatusCode)
	}

	var payload struct {
		Articles []string `json:"articles"`
		Next     string   `json:"next"`
		QueryID  string   `json:"query_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return SearchResult{}, err
	}

	return SearchResult{Articles: payload.Articles, Next: payload.Next, QueryID: payload.QueryID}, nil
}

func (c *Client) Suggest(prefix string, limit int) ([]string, error) {
	q := url.Values{}
	q.Set("q", prefix)
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	endpoint := fmt.Sprintf("%s/search/suggest?%s", c.baseURL, q.Encode())
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
		return nil, fmt.Errorf("recsys suggest failed with status %d", resp.StatusCode)
	}
	var payload struct {
		Suggestions []string `json:"suggestions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload.Suggestions, nil
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
