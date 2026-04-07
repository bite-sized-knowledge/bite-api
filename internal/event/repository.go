package event

import "github.com/jmoiron/sqlx"

type input struct {
	EventUUID     string
	MemberID      *int64
	DeviceID      string
	EventType     string
	ArticleID     *string
	DwellTimeMs   *int64
	ScrollDepth   *int
	Source        string
	Position      *int
	FeedRequestID string
	SessionID     string
	DeviceType    string
	AppVersion    string
	Metadata      []byte
	OccurredAt    string
}

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(event input, upsertHistory bool) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.Exec(`
		INSERT INTO user_events (
			event_uuid,
			member_id,
			device_id,
			event_type,
			article_id,
			dwell_time_ms,
			scroll_depth,
			source,
			position,
			feed_request_id,
			session_id,
			device_type,
			app_version,
			metadata,
			occurred_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.EventUUID,
		event.MemberID,
		event.DeviceID,
		event.EventType,
		event.ArticleID,
		event.DwellTimeMs,
		event.ScrollDepth,
		event.Source,
		event.Position,
		event.FeedRequestID,
		event.SessionID,
		event.DeviceType,
		event.AppVersion,
		event.Metadata,
		event.OccurredAt,
	)
	if err != nil {
		return err
	}

	if upsertHistory && event.MemberID != nil && event.ArticleID != nil && *event.ArticleID != "" {
		if _, err = tx.Exec(`INSERT INTO article_history (member_id, article_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE id = id`, *event.MemberID, *event.ArticleID); err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

func (r *Repository) MergeAnonymous(memberID int64, deviceID string) (int64, error) {
	result, err := r.db.Exec(
		`UPDATE user_events SET member_id = ? WHERE device_id = ? AND member_id IS NULL`,
		memberID, deviceID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
