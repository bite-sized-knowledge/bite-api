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

	var memberIDPtr *int64
	if memberID > 0 {
		memberIDPtr = &memberID
	}

	return s.repo.Create(input{
		EventUUID:     eventUUID,
		MemberID:      memberIDPtr,
		DeviceID:      strings.TrimSpace(req.DeviceID),
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

func (s *Service) MergeAnonymousEvents(memberID int64, deviceID string) (int64, error) {
	if deviceID == "" {
		return 0, fmt.Errorf("%w: deviceId is required", model.ErrBadRequest)
	}
	return s.repo.MergeAnonymous(memberID, deviceID)
}

func shouldUpdateHistory(eventType string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(eventType))
	switch normalized {
	// ARTICLE_IN is the RN app's "user opened reader" signal. bite-web
	// fires ARTICLE_CLICK instead (it opens articles in a new tab, not
	// an in-app reader, so "IN"/"OUT" framing didn't apply cleanly).
	// Both should populate /my/history, so accept either.
	case "ARTICLE_IN", "ARTICLE_OPEN", "ARTICLE_VIEW", "ARTICLE_CLICK", "OPEN", "VIEW", "S_CLICK":
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
