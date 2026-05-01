package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/bite-sized/bite-api/internal/authcookie"
	"github.com/labstack/echo/v4"
)

const (
	HeaderDeviceID   = "X-Device-Id"
	HeaderGuestToken = "X-Guest-Token"
)

type GuestIssuer func(deviceID string) (memberID int64, accessToken, refreshToken string, created bool, err error)

// recsysMigrator: device 가 새로 member 로 mint 된 시점에 recsys 의 bandit state 이관 (fire-and-forget).
// nil 이면 호출 안 함 (테스트 / migration 비활성).
type RecsysMigrator interface {
	MigrateDevice(memberID int64, deviceID string) error
}

// 동시 in-flight migrate 한도. 트래픽 폭증 시 outbound socket 폭증 막는 안전망 —
// 5초 timeout 의 recsys.Client.MigrateDevice 가 한도 도달하면 `select default` 로 즉시 drop.
// 정상 운영에선 lazy guest 발급 자체가 드물고 (FK action 첫 1회) recsys 응답 ms 라 거의 안 닿음.
const migrateMaxInflight = 64

var migrateSem = make(chan struct{}, migrateMaxInflight)

// LazyGuest sits after OptionalJWTAuth on FK-requiring endpoints. If the
// caller is already authenticated it passes through; otherwise it requires
// X-Device-Id, lazily mints a guest member, and exposes the access token via
// X-Guest-Token (refresh token goes into the existing httpOnly cookie).
//
// Members are only minted on actions that require a FK (likes/bookmarks/
// subscribe/interests), keeping bot traffic from polluting the member table.
// Pure browse / event traffic stays anonymous (device_id is the analytics
// key — see internal/event).
//
// 신규 발급 시점에만 recsys.MigrateDevice 를 fire-and-forget 호출 — device_category_bandit
// 의 alpha/beta/impressions/clicks 를 새 member_category_bandit 으로 이관.
func LazyGuest(issue GuestIssuer, migrator RecsysMigrator, refreshExpiry time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if id, ok := c.Get(ContextKeyMemberID).(int64); ok && id != 0 {
				return next(c)
			}
			deviceID := c.Request().Header.Get(HeaderDeviceID)
			if deviceID == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "authentication required: provide bearer token or X-Device-Id header",
				})
			}
			memberID, accessToken, refreshToken, created, err := issue(deviceID)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "failed to issue guest token",
				})
			}
			c.Response().Header().Set(HeaderGuestToken, accessToken)
			authcookie.Set(c, refreshToken, refreshExpiry)
			c.Set(ContextKeyMemberID, memberID)

			if created && migrator != nil {
				select {
				case migrateSem <- struct{}{}:
					go func(mid int64, did string) {
						defer func() { <-migrateSem }()
						if mErr := migrator.MigrateDevice(mid, did); mErr != nil {
							slog.Warn("recsys migrate-device failed",
								slog.Int64("member_id", mid),
								slog.String("device_id", did),
								slog.Any("error", mErr),
							)
						}
					}(memberID, deviceID)
				default:
					slog.Warn("recsys migrate-device dropped (semaphore full)",
						slog.Int64("member_id", memberID),
						slog.String("device_id", deviceID),
						slog.Int("inflight_cap", migrateMaxInflight),
					)
				}
			}
			return next(c)
		}
	}
}
