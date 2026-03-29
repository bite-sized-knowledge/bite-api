package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	jwtpkg "github.com/bite-sized/bite-api/pkg/jwt"
	"github.com/labstack/echo/v4"
)

const (
	ContextKeyMemberID = "member_id"
	ContextKeyClaims   = "member_claims"
)

func JWTAuth(jwtService *jwtpkg.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := bearerToken(c.Request().Header.Get(echo.HeaderAuthorization))
			if token == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "missing authorization token"})
			}

			claims, err := jwtService.ValidateToken(token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "invalid authorization token"})
			}

			setMemberContext(c, claims)
			return next(c)
		}
	}
}

func OptionalJWTAuth(jwtService *jwtpkg.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := bearerToken(c.Request().Header.Get(echo.HeaderAuthorization))
			if token == "" {
				return next(c)
			}

			claims, err := jwtService.ValidateToken(token)
			if err != nil {
				return next(c)
			}

			setMemberContext(c, claims)
			return next(c)
		}
	}
}

func RequireRole(role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := c.Get(ContextKeyClaims).(*jwtpkg.Claims)
			if !ok || claims == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "authentication required"})
			}
			if !strings.EqualFold(claims.Role, role) {
				return c.JSON(http.StatusForbidden, map[string]string{"message": "forbidden"})
			}
			return next(c)
		}
	}
}

func bearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

func setMemberContext(c echo.Context, claims *jwtpkg.Claims) {
	if claims == nil {
		return
	}
	memberID := claims.ID
	if memberID == 0 && claims.Subject != "" {
		if parsed, err := strconv.ParseInt(claims.Subject, 10, 64); err == nil {
			memberID = parsed
		}
	}
	c.Set(ContextKeyClaims, claims)
	c.Set(ContextKeyMemberID, memberID)
}

func CurrentMemberID(c echo.Context) (int64, error) {
	v := c.Get(ContextKeyMemberID)
	memberID, ok := v.(int64)
	if !ok || memberID == 0 {
		return 0, fmt.Errorf("member id not found")
	}
	return memberID, nil
}
