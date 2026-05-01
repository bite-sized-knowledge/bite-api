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

// FeedResult bundles the feed items and the recsys feed_request_id so the
// handler can echo it via the X-Feed-Request-Id response header. Internal —
// the wire response shape stays []article.FeedItem (FeedResponse).
type FeedResult struct {
	Items         []article.FeedItem
	FeedRequestID string
}

func (s *Service) Feed(memberID int64, deviceID string) (FeedResult, error) {
	if memberID > 0 || deviceID != "" {
		res, err := s.recsysClient.GetFeed(memberID, deviceID)
		if err == nil && len(res.Articles) > 0 {
			items, itemErr := s.articleRepo.GetByIDs(memberID, res.Articles)
			if itemErr == nil && len(items) > 0 {
				return FeedResult{Items: items, FeedRequestID: res.FeedRequestID}, nil
			}
		}
	}

	recent, err := s.articleRepo.ListRecent(memberID, 30, "", "", nil)
	if err != nil {
		return FeedResult{}, err
	}
	items := recent.Articles
	rand.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})
	if len(items) > 10 {
		items = items[:10]
	}
	return FeedResult{Items: items}, nil
}
