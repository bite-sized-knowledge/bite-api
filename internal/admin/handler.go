package admin

import (
	"fmt"
	"strconv"

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

func RegisterRoutes(g *echo.Group, h *Handler, middleware ...echo.MiddlewareFunc) {
	if len(middleware) > 0 {
		g.Use(middleware...)
	}
	g.DELETE("/articles/:articleId", h.deleteArticle)
	g.DELETE("/blogs/:blogId", h.deleteBlog)
}

func (h *Handler) deleteArticle(c echo.Context) error {
	if err := h.service.DeleteArticle(c.Param("articleId")); err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, nil)
}

func (h *Handler) deleteBlog(c echo.Context) error {
	blogID, err := strconv.ParseInt(c.Param("blogId"), 10, 64)
	if err != nil {
		return response.Error(c, fmt.Errorf("%w: invalid blogId", model.ErrBadRequest))
	}
	if err := h.service.DeleteBlog(blogID); err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, nil)
}
