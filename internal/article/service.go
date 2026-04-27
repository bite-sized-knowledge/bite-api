package article

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/bite-sized/bite-api/internal/model"
	"github.com/bite-sized/bite-api/internal/recsys"
)

const searchQueryMaxLen = 100

type Service struct {
	repo                 *Repository
	recsysClient         *recsys.Client
	recsysSearchEnabled  bool
}

func NewService(repo *Repository, recsysClient *recsys.Client, recsysSearchEnabled bool) *Service {
	return &Service{
		repo:                repo,
		recsysClient:        recsysClient,
		recsysSearchEnabled: recsysSearchEnabled,
	}
}

func (s *Service) Like(memberID int64, articleID string) error {
	return s.repo.Like(memberID, articleID)
}

func (s *Service) Unlike(memberID int64, articleID string) error {
	return s.repo.Unlike(memberID, articleID)
}

func (s *Service) Uninterest(memberID int64, articleID string) error {
	return s.repo.MarkUninterested(memberID, articleID)
}

func (s *Service) Likes(memberID int64, limit int, from string) (*LikedArticlesPage, error) {
	return s.repo.ListLikes(memberID, normalizeLimit(limit), from)
}

func (s *Service) Bookmarks(memberID int64, limit int, from string) (*BookmarkPage, error) {
	return s.repo.ListBookmarks(memberID, normalizeLimit(limit), from)
}

func (s *Service) Bookmark(memberID int64, articleID string) error {
	return s.repo.Bookmark(memberID, articleID)
}

func (s *Service) Unbookmark(memberID int64, articleID string) error {
	return s.repo.Unbookmark(memberID, articleID)
}

func (s *Service) Share(memberID int64, articleID string) error {
	return s.repo.Share(memberID, articleID)
}

func (s *Service) Recent(memberID int64, limit int, from string, lang string, blogIDs []int64) (*RecentArticlesPage, error) {
	return s.repo.ListRecent(memberID, normalizeLimit(limit), from, lang, blogIDs)
}

func (s *Service) History(memberID int64, limit int, from string) (*ArticleHistoryPage, error) {
	var cursor *int64
	if from != "" {
		parsed, err := strconv.ParseInt(from, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid history cursor", model.ErrBadRequest)
		}
		cursor = &parsed
	}
	return s.repo.ListHistory(memberID, normalizeLimit(limit), cursor)
}

// SearchOptions carries optional filters and the hybrid mode hint for the
// recsys-backed search path. The fallback path (MySQL FULLTEXT) ignores filters
// other than the query — adding them there is a future PR.
type SearchOptions struct {
	MemberID        int64
	CategoryID      *int64
	Lang            string
	BlogID          *int64
	PublishedAfter  *float64
	PublishedBefore *float64
	Mode            string
}

func (s *Service) Search(query string, limit int, from string, opts SearchOptions) (*ArticleSearchPage, error) {
	trimmed, err := validateSearchQuery(query)
	if err != nil {
		return nil, err
	}
	normalized := normalizeLimit(limit)

	if s.recsysSearchEnabled && s.recsysClient != nil {
		page, err := s.searchViaRecsys(trimmed, normalized, from, opts)
		if err == nil {
			return page, nil
		}
		// fallback to MySQL FULLTEXT — observable so the operator notices.
		slog.Warn("recsys search failed, falling back to MySQL FULLTEXT",
			"err", err, "query_len", utf8.RuneCountInString(trimmed))
	}
	return s.repo.Search(trimmed, normalized, from)
}

func (s *Service) searchViaRecsys(query string, limit int, from string, opts SearchOptions) (*ArticleSearchPage, error) {
	req := recsys.SearchRequest{
		Query:           query,
		Limit:           limit,
		From:            from,
		CategoryID:      opts.CategoryID,
		Lang:            opts.Lang,
		BlogID:          opts.BlogID,
		PublishedAfter:  opts.PublishedAfter,
		PublishedBefore: opts.PublishedBefore,
		Mode:            opts.Mode,
	}
	result, err := s.recsysClient.Search(req)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.GetArticlesByIDsPreservingOrder(result.Articles, opts.MemberID)
	if err != nil {
		return nil, err
	}
	return &ArticleSearchPage{Articles: items, Next: result.Next, QueryID: result.QueryID}, nil
}

func validateSearchQuery(query string) (string, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return "", fmt.Errorf("%w: query is required", model.ErrBadRequest)
	}
	if utf8.RuneCountInString(trimmed) > searchQueryMaxLen {
		return "", fmt.Errorf("%w: query exceeds %d characters", model.ErrBadRequest, searchQueryMaxLen)
	}
	return trimmed, nil
}

func (s *Service) ByIDs(memberID int64, req ByIDsRequest) ([]FeedItem, error) {
	return s.repo.GetByIDs(memberID, req.ArticleIDs)
}

// Suggest forwards to recsys-serving's popular-query suggest. Always returns
// an empty slice (never nil) so the JSON response stays well-formed even when
// recsys is unavailable.
func (s *Service) Suggest(prefix string, limit int) []string {
	if s.recsysClient == nil || !s.recsysSearchEnabled {
		return []string{}
	}
	if limit <= 0 || limit > 20 {
		limit = 8
	}
	suggestions, err := s.recsysClient.Suggest(prefix, limit)
	if err != nil {
		slog.Warn("recsys suggest failed", "err", err)
		return []string{}
	}
	if suggestions == nil {
		return []string{}
	}
	return suggestions
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 10
	}
	if limit > 50 {
		return 50
	}
	return limit
}
