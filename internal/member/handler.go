package member

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bite-sized/bite-api/internal/authcookie"
	"github.com/bite-sized/bite-api/internal/middleware"
	"github.com/bite-sized/bite-api/internal/model"
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

func RegisterRoutes(v1 *echo.Group, h *Handler, authMiddleware echo.MiddlewareFunc, optionalAuth echo.MiddlewareFunc, lazyGuest echo.MiddlewareFunc) {
	g := v1.Group("/members")

	regRL := middleware.RateLimit(3, 5)
	g.POST("", h.createGuest, regRL)
	g.POST("/join", h.join, regRL)

	g.PUT("/interests", h.updateInterests, optionalAuth, lazyGuest)

	protected := g.Group("")
	protected.Use(authMiddleware)
	protected.GET("/me", h.getProfile)
	protected.GET("/name/check", h.checkName)
	protected.GET("/interests", h.getInterests)
	protected.DELETE("/:memberId", h.deleteMember)
	protected.PUT("/profile", h.updateProfile)
}

func (h *Handler) createGuest(c echo.Context) error {
	var req CreateGuestRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, err)
	}
	result, err := h.service.CreateGuestMember(req)
	if err != nil {
		return response.Error(c, err)
	}
	authcookie.Set(c, result.Token.RefreshToken, h.refreshExpiry)
	return response.Created(c, result)
}

func (h *Handler) join(c echo.Context) error {
	var req JoinRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, err)
	}
	result, err := h.service.RegisterMember(req)
	if err != nil {
		return response.Error(c, err)
	}
	authcookie.Set(c, result.Token.RefreshToken, h.refreshExpiry)
	return response.Created(c, result)
}

func (h *Handler) checkName(c echo.Context) error {
	duplicated, err := h.service.HasDuplicateName(c.QueryParam("name"))
	if err != nil {
		return response.Error(c, err)
	}
	if duplicated {
		return response.Error(c, fmt.Errorf("%w: name already exists", model.ErrConflict))
	}
	return response.Success(c, nil)
}

func (h *Handler) updateInterests(c echo.Context) error {
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req UpdateInterestsRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, err)
	}
	if err := h.service.UpdateInterests(memberID, req.InterestIDs); err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, nil)
}

func (h *Handler) getInterests(c echo.Context) error {
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	ids, err := h.service.GetInterests(memberID)
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, ids)
}

func (h *Handler) updateProfile(c echo.Context) error {
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, err)
	}
	result, err := h.service.UpdateProfile(memberID, req)
	if err != nil {
		return response.Error(c, err)
	}
	authcookie.Set(c, result.Token.RefreshToken, h.refreshExpiry)
	return response.Success(c, result)
}

func (h *Handler) getProfile(c echo.Context) error {
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	profile, err := h.service.GetProfile(memberID)
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, profile)
}

func (h *Handler) deleteMember(c echo.Context) error {
	currentMemberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	memberID, err := strconv.ParseInt(c.Param("memberId"), 10, 64)
	if err != nil {
		return response.Error(c, err)
	}
	if err := h.service.DeleteMember(currentMemberID, memberID); err != nil {
		return response.Error(c, err)
	}
	return response.NoContent(c)
}
