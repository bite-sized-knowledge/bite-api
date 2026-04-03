package auth

import (
	"fmt"
	"net/http"

	"github.com/bite-sized/bite-api/internal/middleware"
	"github.com/bite-sized/bite-api/pkg/response"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func RegisterRoutes(v1 *echo.Group, h *Handler, oh *OAuthHandler, authMiddleware ...echo.MiddlewareFunc) {
	g := v1.Group("/auth")

	g.POST("/login", h.login)
	g.POST("/refresh", h.refresh)
	g.POST("/password/reset", h.passwordReset)
	g.GET("/email/verify", h.verifyEmail)
	g.POST("/oauth/github", oh.HandleGitHubOAuth)
	g.POST("/oauth/google", oh.HandleGoogleOAuth)

	protected := g.Group("")
	if len(authMiddleware) > 0 {
		protected.Use(authMiddleware...)
	}
	protected.POST("/email/request-verify", h.requestVerifyEmail)
	protected.GET("/email/is-verified", h.isVerified)
	protected.POST("/password/change", h.changePassword)
	protected.POST("/password/match", h.matchPassword)
}

func (h *Handler) login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, err)
	}
	result, err := h.service.Login(req)
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, result)
}

func (h *Handler) refresh(c echo.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, err)
	}
	result, err := h.service.Refresh(req)
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, result)
}

func (h *Handler) requestVerifyEmail(c echo.Context) error {
	var req EmailRequestVerifyRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, err)
	}
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	if err := h.service.SendEmailVerification(req.Email, memberID); err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, nil)
}

func (h *Handler) isVerified(c echo.Context) error {
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	result, err := h.service.IsVerified(c.QueryParam("email"), memberID)
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, result)
}

func (h *Handler) passwordReset(c echo.Context) error {
	var req PasswordResetRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, err)
	}
	if err := h.service.SendPasswordResetEmail(req.Email); err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, nil)
}

func (h *Handler) changePassword(c echo.Context) error {
	var req PasswordChangeRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, err)
	}
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	if err := h.service.ChangePassword(memberID, req); err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, nil)
}

func (h *Handler) matchPassword(c echo.Context) error {
	var req PasswordMatchRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, err)
	}
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	result, err := h.service.MatchPassword(memberID, req)
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, result)
}

func (h *Handler) verifyEmail(c echo.Context) error {
	emailAddress := c.QueryParam("email")
	code := c.QueryParam("code")
	verifyType := c.QueryParam("type")
	if err := h.service.VerifyEmail(emailAddress, code, verifyType); err != nil {
		return c.HTML(http.StatusBadRequest, fmt.Sprintf("<html><body><h1>Verification failed</h1><p>%s</p></body></html>", err.Error()))
	}
	return c.HTML(http.StatusOK, "<html><body><h1>Verification succeeded</h1></body></html>")
}
