package blog

import (
	"fmt"
	"strconv"

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
	g := v1.Group("/blogs")

	if optionalAuth != nil {
		g.GET("", h.list, optionalAuth)
	} else {
		g.GET("", h.list)
	}

	protected := g.Group("")
	if authMiddleware != nil {
		protected.Use(authMiddleware)
	}
	protected.GET("/:blogId", h.get)
	protected.POST("/:blogId/subscribe", h.subscribe)
	protected.DELETE("/:blogId/subscribe", h.unsubscribe)
	protected.GET("/:blogId/articles", h.articles)
}

func (h *Handler) list(c echo.Context) error {
	result, err := h.service.List(optionalMemberID(c))
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, result)
}

func (h *Handler) get(c echo.Context) error {
	blogID, err := parseBlogID(c.Param("blogId"))
	if err != nil {
		return response.Error(c, err)
	}
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	result, err := h.service.Get(blogID, &memberID)
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, result)
}

func (h *Handler) subscribe(c echo.Context) error {
	blogID, err := parseBlogID(c.Param("blogId"))
	if err != nil {
		return response.Error(c, err)
	}
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	if err := h.service.Subscribe(memberID, blogID); err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, nil)
}

func (h *Handler) unsubscribe(c echo.Context) error {
	blogID, err := parseBlogID(c.Param("blogId"))
	if err != nil {
		return response.Error(c, err)
	}
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	if err := h.service.Unsubscribe(memberID, blogID); err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, nil)
}

func (h *Handler) articles(c echo.Context) error {
	blogID, err := parseBlogID(c.Param("blogId"))
	if err != nil {
		return response.Error(c, err)
	}
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	result, err := h.service.Articles(memberID, blogID, queryInt(c, "limit"), c.QueryParam("from"))
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, result)
}

func parseBlogID(raw string) (int64, error) {
	blogID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid blogId", model.ErrBadRequest)
	}
	return blogID, nil
}

func optionalMemberID(c echo.Context) *int64 {
	v := c.Get(middleware.ContextKeyMemberID)
	memberID, ok := v.(int64)
	if !ok || memberID == 0 {
		return nil
	}
	return &memberID
}

func queryInt(c echo.Context, key string) int {
	value := c.QueryParam(key)
	if value == "" {
		return 10
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 10
	}
	return parsed
}
