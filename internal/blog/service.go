package blog

import (
	"fmt"
	"strconv"

	"github.com/bite-sized/bite-api/internal/article"
	"github.com/bite-sized/bite-api/internal/model"
)

type Service struct {
	repo        *Repository
	articleRepo *article.Repository
}

func NewService(repo *Repository, articleRepo *article.Repository) *Service {
	return &Service{repo: repo, articleRepo: articleRepo}
}

func (s *Service) List(memberID *int64) (*ListBlogsResponse, error) {
	rows, err := s.repo.List(memberID)
	if err != nil {
		return nil, err
	}
	blogs := make([]BlogResponse, 0, len(rows))
	for _, r := range rows {
		blogs = append(blogs, BlogResponse{
			ID:           strconv.FormatInt(r.ID, 10),
			Title:        r.Title,
			URL:          r.URL,
			Favicon:      r.Favicon,
			IsSubscribed: r.IsSubscribed,
		})
	}
	return &ListBlogsResponse{Blogs: blogs}, nil
}

func (s *Service) Get(blogID int64, memberID *int64) (*BlogResponse, error) {
	r, err := s.repo.Get(blogID, memberID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, fmt.Errorf("%w: blog not found", model.ErrNotFound)
	}
	return &BlogResponse{
		ID:           strconv.FormatInt(r.ID, 10),
		Title:        r.Title,
		URL:          r.URL,
		Favicon:      r.Favicon,
		IsSubscribed: r.IsSubscribed,
	}, nil
}

func (s *Service) Subscribe(memberID, blogID int64) error {
	exists, err := s.repo.Exists(blogID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%w: invalid blogId", model.ErrBadRequest)
	}
	return s.repo.Subscribe(memberID, blogID)
}

func (s *Service) Unsubscribe(memberID, blogID int64) error {
	exists, err := s.repo.Exists(blogID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%w: invalid blogId", model.ErrBadRequest)
	}
	return s.repo.Unsubscribe(memberID, blogID)
}

func (s *Service) Articles(memberID, blogID int64, limit int, from string) (*BlogArticlesResponse, error) {
	exists, err := s.repo.Exists(blogID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("%w: invalid blogId", model.ErrBadRequest)
	}
	page, err := s.articleRepo.ListByBlog(memberID, blogID, normalizeLimit(limit), from)
	if err != nil {
		return nil, err
	}
	return &BlogArticlesResponse{Articles: page.Articles, Next: page.Next}, nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 || limit > 50 {
		return 10
	}
	return limit
}
