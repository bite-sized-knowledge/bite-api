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
		_ = json.NewEncoder(w).Encode(map[string]any{"articles": []string{"a1", "a2"}})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	articleIDs, err := client.GetFeed(42)
	if err != nil {
		t.Fatalf("GetFeed error: %v", err)
	}
	if len(articleIDs) != 2 || articleIDs[0] != "a1" || articleIDs[1] != "a2" {
		t.Fatalf("unexpected article ids: %#v", articleIDs)
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
		_ = json.NewEncoder(w).Encode(map[string]any{"articles": []string{"z9"}})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	articleIDs, err := client.Search("golang feed")
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(articleIDs) != 1 || articleIDs[0] != "z9" {
		t.Fatalf("unexpected article ids: %#v", articleIDs)
	}
}
