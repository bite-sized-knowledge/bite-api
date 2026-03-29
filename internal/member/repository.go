package member

import (
	"database/sql"

	"github.com/bite-sized/bite-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateGuest(name string) (int64, error) {
	result, err := r.db.Exec(`INSERT INTO member (name, status, role) VALUES (?, 'ACTIVE', 'ROLE_GUEST')`, name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) InterestExists(interestID int64) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM interest WHERE interest_id = ?)`, interestID)
	return exists, err
}

func (r *Repository) AddMemberInterest(memberID, interestID int64) error {
	_, err := r.db.Exec(`INSERT INTO member_interest (member_id, interest_id) VALUES (?, ?)`, memberID, interestID)
	return err
}

func (r *Repository) FindMemberByID(memberID int64) (*model.Member, error) {
	var memberRecord model.Member
	err := r.db.Get(&memberRecord, `SELECT member_id, COALESCE(email, '') AS email, COALESCE(password, '') AS password, COALESCE(name, '') AS name, birth, COALESCE(gender, '') AS gender, COALESCE(status, '') AS status, COALESCE(role, '') AS role, created_at, updated_at FROM member WHERE member_id = ?`, memberID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &memberRecord, nil
}

func (r *Repository) ExistsByEmail(email string) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM member WHERE email = ?)`, email)
	return exists, err
}

func (r *Repository) ExistsByName(name string) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM member WHERE name = ?)`, name)
	return exists, err
}

func (r *Repository) IsEmailVerified(email string, memberID int64) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM email_verify WHERE email = ? AND member_id = ? AND is_verified = true)`, email, memberID)
	return exists, err
}

func (r *Repository) JoinMember(memberID int64, email, hashedPassword string, birth int) error {
	_, err := r.db.Exec(`UPDATE member SET email = ?, password = ?, birth = ?, role = 'ROLE_USER', status = 'ACTIVE', updated_at = CURRENT_TIMESTAMP WHERE member_id = ?`, email, hashedPassword, birth, memberID)
	return err
}

func (r *Repository) DeleteEmailVerification(email string) error {
	_, err := r.db.Exec(`DELETE FROM email_verify WHERE email = ?`, email)
	return err
}

func (r *Repository) SoftDeleteMember(memberID int64) error {
	_, err := r.db.Exec(`UPDATE member SET status = 'DELETED', updated_at = CURRENT_TIMESTAMP WHERE member_id = ?`, memberID)
	return err
}
