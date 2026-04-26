package article

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/bite-sized/bite-api/internal/model"
)

const searchQueryMaxLen = 100

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
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

func (s *Service) Search(query string, limit int, from string) (*ArticleSearchPage, error) {
	trimmed, err := validateSearchQuery(query)
	if err != nil {
		return nil, err
	}
	return s.repo.Search(trimmed, normalizeLimit(limit), from)
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

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 10
	}
	if limit > 50 {
		return 50
	}
	return limit
}
