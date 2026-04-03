package link

import (
	"fmt"
	"net/url"

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
	rawURL, err := s.articleRepo.GetArticleURL(articleID)
	if err != nil {
		return "", err
	}
	if rawURL == "" {
		return "", fmt.Errorf("%w: article not found", model.ErrNotFound)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("%w: invalid article URL", model.ErrBadRequest)
	}
	return rawURL, nil
}
