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

	// 검색 분석 (S_IMP / S_PREVIEW / S_CLICK 시 첨부).
	// query_norm_hash는 서버에서 계산 — 클라이언트가 보내면 recsys-serving의
	// _query_hash와 정렬이 깨질 위험이 있어 의도적으로 받지 않는다.
	QueryID   string `json:"query_id"`
	QueryText string `json:"query_text"`

	LegacyTargetType string `json:"targetType"`
	LegacyTargetID   string `json:"targetId"`
	LegacyEventType  string `json:"eventType"`
}

type MergeRequest struct {
	DeviceID string `json:"deviceId"`
}
