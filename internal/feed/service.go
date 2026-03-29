package feed

import (
	"github.com/bite-sized/bite-api/internal/article"
	"github.com/bite-sized/bite-api/internal/recsys"
)

type Service struct {
	recsysClient *recsys.Client
	articleRepo  *article.Repository
}

func NewService(recsysClient *recsys.Client, articleRepo *article.Repository) *Service {
	return &Service{recsysClient: recsysClient, articleRepo: articleRepo}
}

func (s *Service) Feed(memberID int64) ([]article.FeedItem, error) {
	articleIDs, err := s.recsysClient.GetFeed(memberID)
	if err == nil && len(articleIDs) > 0 {
		items, itemErr := s.articleRepo.GetByIDs(memberID, articleIDs)
		if itemErr == nil && len(items) > 0 {
			return items, nil
		}
	}

	recent, err := s.articleRepo.ListRecent(memberID, 10, "")
	if err != nil {
		return nil, err
	}
	return recent.Articles, nil
}
