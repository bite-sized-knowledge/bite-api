package article

import (
	"fmt"
	"strconv"

	"github.com/bite-sized/bite-api/internal/model"
)

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

func (s *Service) Recent(memberID int64, limit int, from string) (*RecentArticlesPage, error) {
	return s.repo.ListRecent(memberID, normalizeLimit(limit), from)
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
	return s.repo.Search(query, normalizeLimit(limit), from)
}

func (s *Service) ByIDs(memberID int64, req ByIDsRequest) ([]FeedItem, error) {
	return s.repo.GetByIDs(memberID, req.ArticleIDs)
}

func normalizeLimit(limit int) int {
	if limit <= 0 || limit > 50 {
		return 10
	}
	return limit
}
