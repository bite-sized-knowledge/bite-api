// Package authcookie centralizes the HTTP cookie used to carry the refresh
// token. Both the auth and member handler packages issue refresh tokens
// during login and member creation flows; keeping the cookie attributes in
// one place prevents drift (path, SameSite, Secure, etc.).
package authcookie

import (
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

// Name is the HTTP cookie name that stores the refresh token.
const Name = "refresh_token"

// Path scopes the refresh cookie to the auth endpoints. The browser only
// sends the cookie to /v1/auth/*, which keeps it off unrelated API calls.
const Path = "/v1/auth"

// Set writes the refresh token as an httpOnly cookie. Secure is enabled
// outside the local development environment.
func Set(c echo.Context, token string, maxAge time.Duration) {
	c.SetCookie(&http.Cookie{
		Name:     Name,
		Value:    token,
		Path:     Path,
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		Secure:   secureFlag(),
		SameSite: http.SameSiteLaxMode,
	})
}

// Clear expires the refresh cookie so the next request no longer carries it.
func Clear(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     Name,
		Value:    "",
		Path:     Path,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secureFlag(),
		SameSite: http.SameSiteLaxMode,
	})
}

// Get returns the refresh token value from the cookie, or an empty string
// if the cookie is absent.
func Get(c echo.Context) string {
	cookie, err := c.Cookie(Name)
	if err != nil || cookie == nil {
		return ""
	}
	return cookie.Value
}

func secureFlag() bool {
	return os.Getenv("APP_ENV") != "local"
}
