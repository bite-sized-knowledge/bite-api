package link

import (
	"fmt"

	"github.com/bite-sized/bite-api/internal/model"
)

type articleURLFinder interface {
	GetArticleURL(articleID string) (string, error)
}

type Service struct {
	articleRepo articleURLFinder
}

func NewService(articleRepo articleURLFinder) *Service {
	return &Service{articleRepo: articleRepo}
}

func (s *Service) Resolve(articleID string) (string, error) {
	url, err := s.articleRepo.GetArticleURL(articleID)
	if err != nil {
		return "", err
	}
	if url == "" {
		return "", fmt.Errorf("%w: article not found", model.ErrNotFound)
	}
	return url, nil
}
