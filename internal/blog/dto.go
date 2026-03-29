package blog

import "github.com/bite-sized/bite-api/internal/article"

type BlogResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	URL          string `json:"url"`
	Favicon      string `json:"favicon"`
	IsSubscribed bool   `json:"isSubscribed"`
}

type ListBlogsResponse struct {
	Blogs []BlogResponse `json:"blogs"`
}

type BlogArticlesResponse struct {
	Articles []article.FeedItem `json:"articles"`
	Next     string             `json:"next,omitempty"`
}
