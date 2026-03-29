package event

import (
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

func RegisterRoutes(v1 *echo.Group, h *Handler, authMiddleware ...echo.MiddlewareFunc) {
	g := v1.Group("/events")
	if len(authMiddleware) > 0 {
		g.Use(authMiddleware...)
	}
	g.POST("", h.create)
}

func (h *Handler) create(c echo.Context) error {
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req CreateEventRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, err)
	}
	if err := h.service.Create(memberID, req); err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, nil)
}
