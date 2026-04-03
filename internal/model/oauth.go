package model

import "time"

type OAuth struct {
	OAuthID          int64     `db:"oauth_id" json:"oauth_id"`
	MemberID         int64     `db:"member_id" json:"member_id"`
	Provider         string    `db:"provider" json:"provider"`
	ProviderMemberID string    `db:"provider_member_id" json:"provider_member_id"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}
