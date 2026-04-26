package article

import (
	"errors"
	"strings"
	"testing"

	"github.com/bite-sized/bite-api/internal/model"
)

func TestNormalizeLimit(t *testing.T) {
	cases := []struct {
		in   int
		want int
	}{
		{-5, 10},
		{0, 10},
		{1, 1},
		{10, 10},
		{50, 50},
		{51, 50},
		{100, 50},
		{1000, 50},
	}
	for _, c := range cases {
		if got := normalizeLimit(c.in); got != c.want {
			t.Errorf("normalizeLimit(%d) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestValidateSearchQuery(t *testing.T) {
	t.Run("empty rejected", func(t *testing.T) {
		if _, err := validateSearchQuery(""); !errors.Is(err, model.ErrBadRequest) {
			t.Fatalf("expected ErrBadRequest for empty query, got %v", err)
		}
	})

	t.Run("whitespace-only rejected", func(t *testing.T) {
		if _, err := validateSearchQuery("   \t\n"); !errors.Is(err, model.ErrBadRequest) {
			t.Fatalf("expected ErrBadRequest for whitespace, got %v", err)
		}
	})

	t.Run("trims and accepts normal query", func(t *testing.T) {
		got, err := validateSearchQuery("  OpenAI  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "OpenAI" {
			t.Errorf("expected trimmed 'OpenAI', got %q", got)
		}
	})

	t.Run("counts unicode runes, not bytes", func(t *testing.T) {
		// 50 Korean characters = 150 UTF-8 bytes but 50 runes — must accept.
		q := strings.Repeat("가", 50)
		if _, err := validateSearchQuery(q); err != nil {
			t.Fatalf("50 runes should be valid, got %v", err)
		}
	})

	t.Run("rejects over max length", func(t *testing.T) {
		q := strings.Repeat("a", searchQueryMaxLen+1)
		if _, err := validateSearchQuery(q); !errors.Is(err, model.ErrBadRequest) {
			t.Fatalf("expected ErrBadRequest for >max length, got %v", err)
		}
	})

	t.Run("accepts max length boundary", func(t *testing.T) {
		q := strings.Repeat("a", searchQueryMaxLen)
		if _, err := validateSearchQuery(q); err != nil {
			t.Fatalf("max length should be valid, got %v", err)
		}
	})
}

func TestBuildBooleanFulltextQuery(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"OpenAI", `"OpenAI"`},
		{"  spaced  ", `"spaced"`},
		{`malicious "quote" injection`, `"malicious quote injection"`},
		{`+plus -minus *star (paren)`, `"+plus -minus *star (paren)"`},
		{"", `""`},
		{"   ", `""`},
	}
	for _, c := range cases {
		if got := buildBooleanFulltextQuery(c.in); got != c.want {
			t.Errorf("buildBooleanFulltextQuery(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestServiceSearchValidationShortCircuits(t *testing.T) {
	// nil repo — if validation passes, repo would be dereferenced and the
	// test would panic. Validation must reject before reaching the repo.
	s := &Service{repo: nil}

	if _, err := s.Search("", 10, ""); !errors.Is(err, model.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
	if _, err := s.Search(strings.Repeat("a", 200), 10, ""); !errors.Is(err, model.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest for over-long, got %v", err)
	}
}
