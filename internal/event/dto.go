package event

import (
	"encoding/json"
	"time"
)

type CreateEventRequest struct {
	EventUUID        string          `json:"event_uuid"`
	EventType        string          `json:"event_type"`
	ArticleID        string          `json:"article_id"`
	DeviceID         string          `json:"device_id"`
	DwellTimeMs      *int64          `json:"dwell_time_ms"`
	ScrollDepth      *int            `json:"scroll_depth"`
	Source           string          `json:"source"`
	Position         *int            `json:"position"`
	FeedRequestID    string          `json:"feed_request_id"`
	SessionID        string          `json:"session_id"`
	DeviceType       string          `json:"device_type"`
	AppVersion       string          `json:"app_version"`
	Metadata         json.RawMessage `json:"metadata"`
	OccurredAt       *time.Time      `json:"occurred_at"`
	OccurredAtLegacy *time.Time      `json:"occurredAt"`
	Timestamp        *int64          `json:"timestamp"`

	LegacyTargetType string `json:"targetType"`
	LegacyTargetID   string `json:"targetId"`
	LegacyEventType  string `json:"eventType"`
}

type MergeRequest struct {
	DeviceID string `json:"deviceId"`
}
