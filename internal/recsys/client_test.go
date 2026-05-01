package recsys

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetFeedUsesFeedsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/feeds" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("member_id") != "42" {
			t.Fatalf("unexpected member_id: %s", r.URL.Query().Get("member_id"))
		}
		if r.URL.Query().Get("device_id") != "" {
			t.Fatalf("unexpected device_id present: %s", r.URL.Query().Get("device_id"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"articles": []string{"a1", "a2"}})
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	res, err := client.GetFeed(42, "", "")
	if err != nil {
		t.Fatalf("GetFeed error: %v", err)
	}
	if len(res.Articles) != 2 || res.Articles[0] != "a1" || res.Articles[1] != "a2" {
		t.Fatalf("unexpected article ids: %#v", res.Articles)
	}
}

func TestGetFeedAnonymousUsesDeviceID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("device_id") != "abc-123" {
			t.Fatalf("unexpected device_id: %q", q.Get("device_id"))
		}
		if q.Get("member_id") != "" {
			t.Fatalf("unexpected member_id for anonymous: %q", q.Get("member_id"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"articles": []string{}})
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	if _, err := client.GetFeed(0, "abc-123", ""); err != nil {
		t.Fatalf("GetFeed error: %v", err)
	}
}

func TestGetFeedRequiresIdentifier(t *testing.T) {
	client := NewClient("http://unused", "")
	if _, err := client.GetFeed(0, "", ""); err == nil {
		t.Fatal("expected error when both identifiers are empty")
	}
}

func TestSearchUsesQueryParam(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("query") != "golang feed" {
			t.Fatalf("unexpected query: %s", r.URL.Query().Get("query"))
		}
		if r.URL.Query().Get("limit") != "20" {
			t.Fatalf("unexpected limit: %s", r.URL.Query().Get("limit"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"articles": []string{"z9"}, "next": "abc"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	result, err := client.Search(SearchRequest{Query: "golang feed", Limit: 20})
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(result.Articles) != 1 || result.Articles[0] != "z9" {
		t.Fatalf("unexpected article ids: %#v", result.Articles)
	}
	if result.Next != "abc" {
		t.Fatalf("unexpected next: %s", result.Next)
	}
}

func TestSearchPropagatesFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("category_id") != "3" {
			t.Fatalf("expected category_id=3, got %q", q.Get("category_id"))
		}
		if q.Get("lang") != "ko" {
			t.Fatalf("expected lang=ko, got %q", q.Get("lang"))
		}
		if q.Get("mode") != "hybrid" {
			t.Fatalf("expected mode=hybrid, got %q", q.Get("mode"))
		}
		if q.Get("cursor") != "cur123" {
			t.Fatalf("expected cursor=cur123, got %q", q.Get("cursor"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"articles": []string{}})
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	cat := int64(3)
	_, err := client.Search(SearchRequest{
		Query:      "x",
		Limit:      10,
		From:       "cur123",
		CategoryID: &cat,
		Lang:       "ko",
		Mode:       "hybrid",
	})
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
}
