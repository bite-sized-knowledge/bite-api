package feed

import (
	"math/rand"

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

func (s *Service) Feed(memberID int64, deviceID string) ([]article.FeedItem, error) {
	if memberID > 0 || deviceID != "" {
		articleIDs, err := s.recsysClient.GetFeed(memberID, deviceID)
		if err == nil && len(articleIDs) > 0 {
			items, itemErr := s.articleRepo.GetByIDs(memberID, articleIDs)
			if itemErr == nil && len(items) > 0 {
				return items, nil
			}
		}
	}

	recent, err := s.articleRepo.ListRecent(memberID, 30, "", "", nil)
	if err != nil {
		return nil, err
	}
	items := recent.Articles
	rand.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})
	if len(items) > 10 {
		items = items[:10]
	}
	return items, nil
}
