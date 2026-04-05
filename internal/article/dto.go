package article

import "time"

type ByIDsRequest struct {
	ArticleIDs []string `json:"articleIds"`
}

type Category struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Thumbnail string `json:"thumbnail"`
}

type FeedBlogInfo struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Favicon string `json:"favicon"`
}

type FeedItem struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	Keywords     []string     `json:"keywords"`
	URL          string       `json:"url"`
	Thumbnail    string       `json:"thumbnail"`
	LikeCount    int64        `json:"likeCount"`
	ArchiveCount int64        `json:"archiveCount"`
	ShareCount   int64        `json:"shareCount"`
	PublishedAt  *time.Time   `json:"publishedAt"`
	Category     *Category    `json:"category,omitempty"`
	IsLiked      bool         `json:"isLiked"`
	IsArchived   bool         `json:"isArchived"`
	Blog         FeedBlogInfo `json:"blog"`
}

type BookmarkPage struct {
	Articles []FeedItem `json:"articles"`
	Next     string     `json:"next,omitempty"`
}

type RecentArticlesPage struct {
	Articles []FeedItem `json:"articles"`
	Next     string     `json:"next,omitempty"`
}

type ArticleHistoryPage struct {
	Articles []FeedItem `json:"articles"`
	Next     *int64     `json:"next,omitempty"`
}

type ArticleSearchContainer struct {
	Articles []FeedItem `json:"articles"`
}
