package event

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/bite-sized/bite-api/internal/model"
)

// recsysFeedbackPoster is the minimal surface event.Service needs from
// recsys.Client. Defined here so tests can stub it without importing recsys.
// memberID > 0 = 회원, deviceID = 비회원 (둘 중 하나는 보장).
type recsysFeedbackPoster interface {
	PostFeedback(memberID int64, deviceID, articleID, eventType string) error
}

// recsysFeedbackEvents are the bite-api event types that map to bandit/user_vector
// reward signals on recsys-serving. Any other event_type is dropped (no HTTP call).
// Mapping → recsys: ARTICLE_IN→article_in, LIKE→like, ARCHIVE→archive,
// SHARE→share, UNINTEREST→uninterest. ARTICLE_CLICK 은 RN/web 둘 다 클릭이라
// article_in 으로 normalize.
var recsysFeedbackEventMap = map[string]string{
	"ARTICLE_IN":    "article_in",
	"ARTICLE_CLICK": "article_in",
	"ARTICLE_OPEN":  "article_in",
	"LIKE":          "like",
	"ARCHIVE":       "archive",
	"SHARE":         "share",
	"UNINTEREST":    "uninterest",
}

const queryTextMaxLen = 200

// normalizeAndHashQuery는 recsys-serving의 _query_hash와 동일 알고리즘.
// (lower + strip 후 sha1 → hex 처음 12자.)
func normalizeAndHashQuery(query string) string {
	normalized := strings.ToLower(strings.TrimSpace(query))
	if normalized == "" {
		return ""
	}
	sum := sha1.Sum([]byte(normalized))
	return hex.EncodeToString(sum[:])[:12]
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

type Service struct {
	repo   *Repository
	recsys recsysFeedbackPoster
}

func NewService(repo *Repository, recsys recsysFeedbackPoster) *Service {
	return &Service{repo: repo, recsys: recsys}
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

	queryText := truncateRunes(strings.TrimSpace(req.QueryText), queryTextMaxLen)
	var queryNormHash string
	if queryText != "" {
		queryNormHash = normalizeAndHashQuery(queryText)
	}

	if err := s.repo.Create(input{
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
		QueryID:       strings.TrimSpace(req.QueryID),
		QueryText:     queryText,
		QueryNormHash: queryNormHash,
	}, shouldUpdateHistory(eventType)); err != nil {
		return err
	}

	// fire-and-forget recsys feedback (실시간 reward). 회원/비회원 둘 다 — 비회원은 device_id.
	s.fireRecsysFeedback(memberIDPtr, strings.TrimSpace(req.DeviceID), articleIDPtr, eventType)
	return nil
}

func (s *Service) fireRecsysFeedback(memberID *int64, deviceID string, articleID *string, eventType string) {
	if s.recsys == nil || articleID == nil {
		return
	}
	// 회원/비회원 모두 article_id 필수. 식별자는 member_id 또는 device_id 중 하나는 있어야.
	hasMember := memberID != nil && *memberID > 0
	if !hasMember && deviceID == "" {
		return
	}
	mapped, ok := recsysFeedbackEventMap[strings.ToUpper(strings.TrimSpace(eventType))]
	if !ok {
		return
	}
	var mid int64
	if hasMember {
		mid = *memberID
	}
	aid := *articleID
	did := deviceID
	go func() {
		if err := s.recsys.PostFeedback(mid, did, aid, mapped); err != nil {
			slog.Warn("recsys feedback post failed",
				slog.Int64("member_id", mid),
				slog.String("device_id", did),
				slog.String("article_id", aid),
				slog.String("event_type", mapped),
				slog.Any("error", err),
			)
		}
	}()
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
