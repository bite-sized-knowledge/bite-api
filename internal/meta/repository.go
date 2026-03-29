package meta

import "github.com/jmoiron/sqlx"

type interestRow struct {
	ID        int64  `db:"id"`
	Name      string `db:"name"`
	Image     string `db:"image"`
	Thumbnail string `db:"thumbnail"`
}

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListInterests() ([]interestRow, error) {
	rows := make([]interestRow, 0)
	err := r.db.Select(&rows, `
		SELECT interest_id AS id, name, image, thumbnail
		FROM interest
		ORDER BY interest_id ASC`)
	return rows, err
}
