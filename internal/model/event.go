package model

import (
	"encoding/json"
	"time"
)

type UserEvent struct {
	ID            int64           `db:"id" json:"id"`
	EventUUID     string          `db:"event_uuid" json:"event_uuid"`
	MemberID      int64           `db:"member_id" json:"member_id"`
	EventType     string          `db:"event_type" json:"event_type"`
	ArticleID     string          `db:"article_id" json:"article_id"`
	DwellTimeMs   int64           `db:"dwell_time_ms" json:"dwell_time_ms"`
	ScrollDepth   int             `db:"scroll_depth" json:"scroll_depth"`
	Source        string          `db:"source" json:"source"`
	Position      int             `db:"position" json:"position"`
	FeedRequestID string          `db:"feed_request_id" json:"feed_request_id"`
	SessionID     string          `db:"session_id" json:"session_id"`
	DeviceType    string          `db:"device_type" json:"device_type"`
	AppVersion    string          `db:"app_version" json:"app_version"`
	Metadata      json.RawMessage `db:"metadata" json:"metadata"`
	OccurredAt    time.Time       `db:"occurred_at" json:"occurred_at"`
	ReceivedAt    time.Time       `db:"received_at" json:"received_at"`
}
