package link

import (
	"net/http"

	"github.com/bite-sized/bite-api/pkg/response"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func RegisterRoutes(v1 *echo.Group, h *Handler) {
	g := v1.Group("/links")
	g.GET("/:articleId", h.redirect)
}

func (h *Handler) redirect(c echo.Context) error {
	url, err := h.service.Resolve(c.Param("articleId"))
	if err != nil {
		return response.Error(c, err)
	}
	return c.Redirect(http.StatusFound, url)
}
