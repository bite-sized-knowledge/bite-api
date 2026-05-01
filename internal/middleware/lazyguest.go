package middleware

import (
	"net/http"
	"time"

	"github.com/bite-sized/bite-api/internal/authcookie"
	"github.com/labstack/echo/v4"
)

const (
	HeaderDeviceID   = "X-Device-Id"
	HeaderGuestToken = "X-Guest-Token"
)

type GuestIssuer func(deviceID string) (memberID int64, accessToken, refreshToken string, err error)

// LazyGuest sits after OptionalJWTAuth on FK-requiring endpoints. If the
// caller is already authenticated it passes through; otherwise it requires
// X-Device-Id, lazily mints a guest member, and exposes the access token via
// X-Guest-Token (refresh token goes into the existing httpOnly cookie).
//
// Members are only minted on actions that require a FK (likes/bookmarks/
// subscribe/interests), keeping bot traffic from polluting the member table.
// Pure browse / event traffic stays anonymous (device_id is the analytics
// key — see internal/event).
func LazyGuest(issue GuestIssuer, refreshExpiry time.Duration) echo.MiddlewareFunc {
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
			memberID, accessToken, refreshToken, err := issue(deviceID)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "failed to issue guest token",
				})
			}
			c.Response().Header().Set(HeaderGuestToken, accessToken)
			authcookie.Set(c, refreshToken, refreshExpiry)
			c.Set(ContextKeyMemberID, memberID)
			return next(c)
		}
	}
}
