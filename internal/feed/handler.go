package feed

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
	g := v1.Group("/feed")
	if len(authMiddleware) > 0 {
		g.Use(authMiddleware...)
	}
	g.GET("", h.feed)
}

func (h *Handler) feed(c echo.Context) error {
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	items, err := h.service.Feed(memberID)
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, items)
}
