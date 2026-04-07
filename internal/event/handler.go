package event

import (
	"fmt"
	"strings"

	"github.com/bite-sized/bite-api/internal/middleware"
	"github.com/bite-sized/bite-api/internal/model"
	"github.com/bite-sized/bite-api/pkg/response"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func RegisterRoutes(v1 *echo.Group, h *Handler, authMiddleware echo.MiddlewareFunc, optionalAuth echo.MiddlewareFunc) {
	g := v1.Group("/events")
	g.POST("", h.create, optionalAuth)
	g.POST("/merge", h.merge, authMiddleware)
}

func (h *Handler) create(c echo.Context) error {
	memberID, _ := middleware.CurrentMemberID(c) // 0 if anonymous
	var req CreateEventRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, err)
	}
	if memberID == 0 && strings.TrimSpace(req.DeviceID) == "" {
		return response.Error(c, fmt.Errorf("%w: device_id is required for anonymous events", model.ErrBadRequest))
	}
	if err := h.service.Create(memberID, req); err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, nil)
}

func (h *Handler) merge(c echo.Context) error {
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req MergeRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, err)
	}
	count, err := h.service.MergeAnonymousEvents(memberID, req.DeviceID)
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, map[string]int64{"merged": count})
}
