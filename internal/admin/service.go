package admin

import "github.com/bite-sized/bite-api/internal/article"

type blogDeleter interface {
	DeleteCascade(blogID int64) error
}

type Service struct {
	articleRepo *article.Repository
	blogRepo    blogDeleter
}

func NewService(articleRepo *article.Repository, blogRepo blogDeleter) *Service {
	return &Service{articleRepo: articleRepo, blogRepo: blogRepo}
}

func (s *Service) DeleteArticle(articleID string) error {
	return s.articleRepo.DeleteArticle(articleID)
}

func (s *Service) DeleteBlog(blogID int64) error {
	return s.blogRepo.DeleteCascade(blogID)
}
