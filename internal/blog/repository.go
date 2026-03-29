package blog

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type row struct {
	ID           int64  `db:"id"`
	Title        string `db:"title"`
	URL          string `db:"url"`
	Favicon      string `db:"favicon"`
	IsSubscribed bool   `db:"is_subscribed"`
}

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(memberID *int64) ([]row, error) {
	rows := make([]row, 0)
	if memberID == nil {
		err := r.db.Select(&rows, `
			SELECT b.blog_id AS id, b.title, b.url, b.favicon, false AS is_subscribed
			FROM blog b
			ORDER BY b.blog_id ASC`)
		return rows, err
	}

	err := r.db.Select(&rows, `
		SELECT b.blog_id AS id, b.title, b.url, b.favicon,
		       EXISTS(
				SELECT 1
				FROM blog_subscribe bs
				WHERE bs.blog_id = b.blog_id
				  AND bs.member_id = ?
				  AND bs.is_deleted = false
			) AS is_subscribed
		FROM blog b
		ORDER BY b.blog_id ASC`, *memberID)
	return rows, err
}

func (r *Repository) Get(blogID int64, memberID *int64) (*row, error) {
	out := row{}
	if memberID == nil {
		err := r.db.Get(&out, `
			SELECT b.blog_id AS id, b.title, b.url, b.favicon, false AS is_subscribed
			FROM blog b
			WHERE b.blog_id = ?`, blogID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		return &out, nil
	}

	err := r.db.Get(&out, `
		SELECT b.blog_id AS id, b.title, b.url, b.favicon,
		       EXISTS(
				SELECT 1
				FROM blog_subscribe bs
				WHERE bs.blog_id = b.blog_id
				  AND bs.member_id = ?
				  AND bs.is_deleted = false
			) AS is_subscribed
		FROM blog b
		WHERE b.blog_id = ?`, *memberID, blogID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *Repository) Exists(blogID int64) (bool, error) {
	var count int64
	err := r.db.Get(&count, `SELECT COUNT(*) FROM blog WHERE blog_id = ?`, blogID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) Subscribe(memberID, blogID int64) error {
	_, err := r.db.Exec(`
		INSERT INTO blog_subscribe (blog_id, member_id, is_deleted)
		VALUES (?, ?, false)
		ON DUPLICATE KEY UPDATE is_deleted = false, updated_at = CURRENT_TIMESTAMP`, blogID, memberID)
	return err
}

func (r *Repository) Unsubscribe(memberID, blogID int64) error {
	_, err := r.db.Exec(`
		UPDATE blog_subscribe
		SET is_deleted = true, updated_at = CURRENT_TIMESTAMP
		WHERE blog_id = ? AND member_id = ?`, blogID, memberID)
	return err
}

func (r *Repository) DeleteCascade(blogID int64) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Exec(`DELETE FROM blog_subscribe WHERE blog_id = ?`, blogID); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM article WHERE blog_id = ?`, blogID); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM blog WHERE blog_id = ?`, blogID); err != nil {
		return err
	}

	err = tx.Commit()
	return err
}
