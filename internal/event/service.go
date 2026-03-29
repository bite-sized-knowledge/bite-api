package event

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/bite-sized/bite-api/internal/model"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(memberID int64, req CreateEventRequest) error {
	eventType := strings.TrimSpace(req.EventType)
	if eventType == "" {
		eventType = strings.TrimSpace(req.LegacyEventType)
	}
	if eventType == "" {
		return fmt.Errorf("%w: eventType is required", model.ErrBadRequest)
	}

	articleID := strings.TrimSpace(req.ArticleID)
	if articleID == "" && strings.EqualFold(strings.TrimSpace(req.LegacyTargetType), "ARTICLE") {
		articleID = strings.TrimSpace(req.LegacyTargetID)
	}

	eventUUID := strings.TrimSpace(req.EventUUID)
	if eventUUID == "" {
		eventUUID = newEventUUID()
	}

	occurredAt := time.Now().UTC()
	if req.OccurredAt != nil {
		occurredAt = req.OccurredAt.UTC()
	} else if req.OccurredAtLegacy != nil {
		occurredAt = req.OccurredAtLegacy.UTC()
	} else if req.Timestamp != nil && *req.Timestamp > 0 {
		occurredAt = time.UnixMilli(*req.Timestamp).UTC()
	}

	var articleIDPtr *string
	if articleID != "" {
		articleIDPtr = &articleID
	}

	return s.repo.Create(input{
		EventUUID:     eventUUID,
		MemberID:      memberID,
		EventType:     eventType,
		ArticleID:     articleIDPtr,
		DwellTimeMs:   req.DwellTimeMs,
		ScrollDepth:   req.ScrollDepth,
		Source:        strings.TrimSpace(req.Source),
		Position:      req.Position,
		FeedRequestID: strings.TrimSpace(req.FeedRequestID),
		SessionID:     strings.TrimSpace(req.SessionID),
		DeviceType:    strings.TrimSpace(req.DeviceType),
		AppVersion:    strings.TrimSpace(req.AppVersion),
		Metadata:      req.Metadata,
		OccurredAt:    occurredAt.Format("2006-01-02 15:04:05"),
	}, shouldUpdateHistory(eventType))
}

func shouldUpdateHistory(eventType string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(eventType))
	switch normalized {
	case "ARTICLE_IN", "ARTICLE_OPEN", "ARTICLE_VIEW", "OPEN", "VIEW":
		return true
	default:
		return false
	}
}

func newEventUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
