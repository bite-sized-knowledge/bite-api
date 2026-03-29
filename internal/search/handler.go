package search

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func RegisterRoutes(v1 *echo.Group, h *Handler, authMiddleware ...echo.MiddlewareFunc) {
	g := v1.Group("/search")
	if len(authMiddleware) > 0 {
		g.Use(authMiddleware...)
	}
	g.GET("", h.notImplemented)
}

func (h *Handler) notImplemented(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"message": "not implemented"})
}
