package auth

import (
	"time"

	"github.com/bite-sized/bite-api/pkg/response"
	"github.com/labstack/echo/v4"
)

type OAuthHandler struct {
	service       *OAuthService
	refreshExpiry time.Duration
}

func NewOAuthHandler(service *OAuthService, refreshExpiry time.Duration) *OAuthHandler {
	return &OAuthHandler{service: service, refreshExpiry: refreshExpiry}
}

type oauthRequest struct {
	Code string `json:"code"`
}

func (h *OAuthHandler) HandleGitHubOAuth(c echo.Context) error {
	var req oauthRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, err)
	}
	result, err := h.service.HandleGitHubLogin(req.Code)
	if err != nil {
		return response.Error(c, err)
	}
	setRefreshTokenCookie(c, result.RefreshToken, h.refreshExpiry)
	return response.Success(c, result)
}

func (h *OAuthHandler) HandleGoogleOAuth(c echo.Context) error {
	var req oauthRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, err)
	}
	result, err := h.service.HandleGoogleLogin(req.Code)
	if err != nil {
		return response.Error(c, err)
	}
	setRefreshTokenCookie(c, result.RefreshToken, h.refreshExpiry)
	return response.Success(c, result)
}
