package auth

import (
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

const refreshTokenCookieName = "refresh_token"

func setRefreshTokenCookie(c echo.Context, token string, maxAge time.Duration) {
	secure := os.Getenv("APP_ENV") != "local"
	sameSite := http.SameSiteLaxMode

	c.SetCookie(&http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    token,
		Path:     "/v1/auth",
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

func clearRefreshTokenCookie(c echo.Context) {
	secure := os.Getenv("APP_ENV") != "local"

	c.SetCookie(&http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/v1/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func getRefreshTokenFromCookie(c echo.Context) string {
	cookie, err := c.Cookie(refreshTokenCookieName)
	if err != nil || cookie == nil {
		return ""
	}
	return cookie.Value
}
