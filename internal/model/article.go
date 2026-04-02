package model

import "time"

type Article struct {
	ArticleID     string     `db:"article_id" json:"article_id"`
	BlogID        int64      `db:"blog_id" json:"blog_id"`
	URL           string     `db:"url" json:"url"`
	Title         string     `db:"title" json:"title"`
	Thumbnail     string     `db:"thumbnail" json:"thumbnail"`
	Description   string     `db:"description" json:"description"`
	Keywords      string     `db:"keywords" json:"keywords"`
	CategoryID    *int64     `db:"category_id" json:"category_id,omitempty"`
	Content       string     `db:"content" json:"content"`
	ContentLength int        `db:"content_length" json:"content_length"`
	Lang          string     `db:"lang" json:"lang"`
	LikeCount     int64      `db:"like_count" json:"like_count"`
	ShareCount    int64      `db:"share_count" json:"share_count"`
	BookmarkCount int64      `db:"bookmark_count" json:"bookmark_count"`
	PublishedAt   *time.Time `db:"published_at" json:"published_at,omitempty"`
	SortKey       string     `db:"sort_key" json:"sort_key"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}
