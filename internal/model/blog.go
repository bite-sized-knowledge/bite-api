package model

import "time"

type Blog struct {
	BlogID     int64     `db:"blog_id" json:"blog_id"`
	PlatformID int64     `db:"platform_id" json:"platform_id"`
	Title      string    `db:"title" json:"title"`
	URL        string    `db:"url" json:"url"`
	RSSURL     string    `db:"rss_url" json:"rss_url"`
	Favicon    string    `db:"favicon" json:"favicon"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}
