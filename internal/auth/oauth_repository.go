package auth

import (
	"database/sql"

	"github.com/bite-sized/bite-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type OAuthRepository struct {
	db *sqlx.DB
}

func NewOAuthRepository(db *sqlx.DB) *OAuthRepository {
	return &OAuthRepository{db: db}
}

func (r *OAuthRepository) FindByProviderAndProviderMemberID(provider, providerMemberID string) (*model.OAuth, error) {
	var record model.OAuth
	err := r.db.Get(&record, `SELECT oauth_id, member_id, provider, provider_member_id, created_at, updated_at FROM oauth WHERE provider = ? AND provider_member_id = ?`, provider, providerMemberID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func (r *OAuthRepository) Create(memberID int64, provider, providerMemberID string) error {
	_, err := r.db.Exec(`INSERT INTO oauth (member_id, provider, provider_member_id) VALUES (?, ?, ?)`, memberID, provider, providerMemberID)
	return err
}
