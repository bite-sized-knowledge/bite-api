package auth

import (
	"github.com/bite-sized/bite-api/pkg/response"
	"github.com/labstack/echo/v4"
)

type OAuthHandler struct {
	service *OAuthService
}

func NewOAuthHandler(service *OAuthService) *OAuthHandler {
	return &OAuthHandler{service: service}
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
	return response.Success(c, result)
}
