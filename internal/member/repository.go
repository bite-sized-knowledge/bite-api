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

func (r *Repository) DB() *sqlx.DB {
	return r.db
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

func (r *Repository) AllInterestsExist(ids []int64) (bool, error) {
	if len(ids) == 0 {
		return false, nil
	}
	query, args, err := sqlx.In(`SELECT COUNT(*) FROM interest WHERE interest_id IN (?)`, ids)
	if err != nil {
		return false, err
	}
	var count int
	if err := r.db.Get(&count, r.db.Rebind(query), args...); err != nil {
		return false, err
	}
	return count == len(ids), nil
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

func (r *Repository) IsEmailVerified(email string) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM email_verify WHERE email = ? AND is_verified = true)`, email)
	return exists, err
}

func (r *Repository) CreateMember(email, hashedPassword string, birth int, name string) (int64, error) {
	result, err := r.db.Exec(`INSERT INTO member (email, password, birth, name, status, role) VALUES (?, ?, ?, ?, 'ACTIVE', 'ROLE_USER')`, email, hashedPassword, birth, name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) DeleteEmailVerification(email string) error {
	_, err := r.db.Exec(`DELETE FROM email_verify WHERE email = ?`, email)
	return err
}

func (r *Repository) SoftDeleteMember(memberID int64) error {
	_, err := r.db.Exec(`UPDATE member SET status = 'DELETED', updated_at = CURRENT_TIMESTAMP WHERE member_id = ?`, memberID)
	return err
}

func (r *Repository) ReplaceInterests(memberID int64, interestIDs []int64) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM member_interest WHERE member_id = ?`, memberID); err != nil {
		return err
	}
	for _, id := range interestIDs {
		if _, err := tx.Exec(`INSERT INTO member_interest (member_id, interest_id) VALUES (?, ?)`, memberID, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) GetMemberInterestIDs(memberID int64) ([]int64, error) {
	var ids []int64
	err := r.db.Select(&ids, `SELECT interest_id FROM member_interest WHERE member_id = ? ORDER BY interest_id`, memberID)
	return ids, err
}
