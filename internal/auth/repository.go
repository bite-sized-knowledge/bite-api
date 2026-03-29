package auth

import (
	"database/sql"
	"time"

	"github.com/bite-sized/bite-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type EmailVerify struct {
	ID         int64     `db:"email_verify_id"`
	Email      string    `db:"email"`
	VerifyCode string    `db:"verify_code"`
	IsVerified bool      `db:"is_verified"`
	MemberID   *int64    `db:"member_id"`
	ExpiredAt  time.Time `db:"expired_at"`
	Type       string    `db:"type"`
}

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindMemberByEmail(email string) (*model.Member, error) {
	var member model.Member
	err := r.db.Get(&member, `SELECT member_id, COALESCE(email, '') AS email, COALESCE(password, '') AS password, COALESCE(name, '') AS name, birth, COALESCE(gender, '') AS gender, COALESCE(status, '') AS status, COALESCE(role, '') AS role, created_at, updated_at FROM member WHERE email = ?`, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &member, nil
}

func (r *Repository) FindMemberByID(memberID int64) (*model.Member, error) {
	var member model.Member
	err := r.db.Get(&member, `SELECT member_id, COALESCE(email, '') AS email, COALESCE(password, '') AS password, COALESCE(name, '') AS name, birth, COALESCE(gender, '') AS gender, COALESCE(status, '') AS status, COALESCE(role, '') AS role, created_at, updated_at FROM member WHERE member_id = ?`, memberID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &member, nil
}

func (r *Repository) FindEmailVerify(email, verifyType string) (*EmailVerify, error) {
	var verify EmailVerify
	err := r.db.Get(&verify, `SELECT email_verify_id, email, verify_code, is_verified, member_id, expired_at, type FROM email_verify WHERE email = ? AND type = ?`, email, verifyType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &verify, nil
}

func (r *Repository) UpsertEmailVerify(email, code string, memberID int64, verifyType string, expiredAt time.Time) error {
	_, err := r.db.Exec(`
		INSERT INTO email_verify (email, verify_code, is_verified, member_id, expired_at, type)
		VALUES (?, ?, false, ?, ?, ?)
		ON DUPLICATE KEY UPDATE verify_code = VALUES(verify_code), is_verified = false, member_id = VALUES(member_id), expired_at = VALUES(expired_at), type = VALUES(type)
	`, email, code, memberID, expiredAt, verifyType)
	return err
}

func (r *Repository) MarkEmailVerified(email, verifyType string) error {
	_, err := r.db.Exec(`UPDATE email_verify SET is_verified = true, updated_at = CURRENT_TIMESTAMP WHERE email = ? AND type = ?`, email, verifyType)
	return err
}

func (r *Repository) DeleteEmailVerify(email, verifyType string) error {
	_, err := r.db.Exec(`DELETE FROM email_verify WHERE email = ? AND type = ?`, email, verifyType)
	return err
}

func (r *Repository) UpdateMemberPassword(memberID int64, hashedPassword string) error {
	_, err := r.db.Exec(`UPDATE member SET password = ?, updated_at = CURRENT_TIMESTAMP WHERE member_id = ?`, hashedPassword, memberID)
	return err
}
