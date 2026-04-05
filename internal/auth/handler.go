package auth

import (
	"html"
	"net/http"
	"time"

	"github.com/bite-sized/bite-api/internal/authcookie"
	"github.com/bite-sized/bite-api/internal/middleware"
	"github.com/bite-sized/bite-api/pkg/response"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service       *Service
	refreshExpiry time.Duration
}

func NewHandler(service *Service, refreshExpiry time.Duration) *Handler {
	return &Handler{service: service, refreshExpiry: refreshExpiry}
}

func RegisterRoutes(v1 *echo.Group, h *Handler, oh *OAuthHandler, authMiddleware ...echo.MiddlewareFunc) {
	g := v1.Group("/auth")

	g.POST("/login", h.login)
	g.POST("/refresh", h.refresh)
	g.POST("/logout", h.logout)
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
	authcookie.Set(c, result.Token.RefreshToken, h.refreshExpiry)
	return response.Success(c, result)
}

func (h *Handler) refresh(c echo.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, err)
	}
	// Prefer cookie, fall back to body for backward compatibility
	if cookieToken := authcookie.Get(c); cookieToken != "" {
		req.RefreshToken = cookieToken
	}
	if req.RefreshToken == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "missing refresh token"})
	}
	result, err := h.service.Refresh(req)
	if err != nil {
		authcookie.Clear(c)
		return response.Error(c, err)
	}
	authcookie.Set(c, result.RefreshToken, h.refreshExpiry)
	return response.Success(c, result)
}

func (h *Handler) logout(c echo.Context) error {
	authcookie.Clear(c)
	return response.Success(c, nil)
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
	// Always return success to prevent account enumeration
	_ = h.service.SendPasswordResetEmail(req.Email)
	return response.Success(c, map[string]string{"message": "If account exists, reset email was sent"})
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

func (h *Handler) verifyEmail(c echo.Context) error {
	emailAddress := c.QueryParam("email")
	code := c.QueryParam("code")
	verifyType := c.QueryParam("type")
	if err := h.service.VerifyEmail(emailAddress, code, verifyType); err != nil {
		return c.HTML(http.StatusBadRequest, "<html><body><h1>Verification failed</h1><p>"+html.EscapeString(err.Error())+"</p></body></html>")
	}
	return c.HTML(http.StatusOK, "<html><body><h1>Verification succeeded</h1></body></html>")
}
