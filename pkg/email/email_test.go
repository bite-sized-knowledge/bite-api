package email

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendVerificationEmail(t *testing.T) {
	var authHeader string
	var payload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emails" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		authHeader = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"email_123"}`))
	}))
	defer server.Close()

	client := NewClientWithBaseURL("re_test", "Bite <noreply@bite-sized.xyz>", server.URL, server.Client())
	if err := client.SendVerificationEmail("user@example.com", "123456", "https://api.bite-sized.xyz/v1/auth/email/verify?code=123456"); err != nil {
		t.Fatalf("SendVerificationEmail returned error: %v", err)
	}

	if authHeader != "Bearer re_test" {
		t.Fatalf("unexpected auth header: %s", authHeader)
	}
	if payload["from"] != "Bite <noreply@bite-sized.xyz>" {
		t.Fatalf("unexpected from: %v", payload["from"])
	}
	if payload["subject"] == "" {
		t.Fatal("expected subject to be set")
	}
}

func TestSendFailsWhenUnconfigured(t *testing.T) {
	client := NewClient("", "")
	if err := client.SendTemporaryPassword("user@example.com", "temp-123"); err == nil {
		t.Fatal("expected configuration error")
	}
}
