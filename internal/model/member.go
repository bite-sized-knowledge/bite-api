package model

import "time"

type Member struct {
	MemberID  int64     `db:"member_id" json:"member_id"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password" json:"-"`
	Name      string    `db:"name" json:"name"`
	Birth     *int      `db:"birth" json:"birth,omitempty"`
	Gender    string    `db:"gender" json:"gender"`
	Status    string    `db:"status" json:"status"`
	Role      string    `db:"role" json:"role"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
