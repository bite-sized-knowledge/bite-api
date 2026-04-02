package model

import "time"

type Interest struct {
	InterestID int64     `db:"interest_id" json:"interest_id"`
	Name       string    `db:"name" json:"name"`
	Image      string    `db:"image" json:"image"`
	Thumbnail  string    `db:"thumbnail" json:"thumbnail"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}
