package article

import (
	"strconv"

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

func RegisterRoutes(v1 *echo.Group, h *Handler, authMiddleware echo.MiddlewareFunc, optionalAuth echo.MiddlewareFunc) {
	g := v1.Group("/articles")

	// Public endpoints (optional auth)
	g.GET("/recent", h.recent, optionalAuth)
	g.GET("/search", h.search)

	// Protected endpoints
	protected := g.Group("")
	protected.Use(authMiddleware)
	protected.POST("/:articleId/likes", h.like)
	protected.DELETE("/:articleId/likes", h.unlike)
	protected.POST("/:articleId/uninterests", h.uninterest)
	protected.GET("/bookmarks", h.bookmarks)
	protected.POST("/:articleId/bookmarks", h.bookmark)
	protected.DELETE("/:articleId/bookmarks", h.unbookmark)
	protected.POST("/:articleId/shares", h.share)
	protected.GET("/history", h.history)
	protected.POST("/by-ids", h.byIDs)
}

func (h *Handler) like(c echo.Context) error {
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	if err := h.service.Like(memberID, c.Param("articleId")); err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, nil)
}

func (h *Handler) unlike(c echo.Context) error {
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	if err := h.service.Unlike(memberID, c.Param("articleId")); err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, nil)
}

func (h *Handler) uninterest(c echo.Context) error {
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	if err := h.service.Uninterest(memberID, c.Param("articleId")); err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, nil)
}

func (h *Handler) bookmarks(c echo.Context) error {
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	result, err := h.service.Bookmarks(memberID, queryInt(c, "limit"), c.QueryParam("from"))
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, result)
}

func (h *Handler) bookmark(c echo.Context) error {
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	if err := h.service.Bookmark(memberID, c.Param("articleId")); err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, nil)
}

func (h *Handler) unbookmark(c echo.Context) error {
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	if err := h.service.Unbookmark(memberID, c.Param("articleId")); err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, nil)
}

func (h *Handler) share(c echo.Context) error {
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	if err := h.service.Share(memberID, c.Param("articleId")); err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, nil)
}

func (h *Handler) recent(c echo.Context) error {
	memberID, _ := middleware.CurrentMemberID(c)
	result, err := h.service.Recent(memberID, queryInt(c, "limit"), c.QueryParam("from"))
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, result)
}

func (h *Handler) history(c echo.Context) error {
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	result, err := h.service.History(memberID, queryInt(c, "limit"), c.QueryParam("from"))
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, result)
}

func (h *Handler) search(c echo.Context) error {
	result, err := h.service.Search(c.QueryParam("query"))
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, result)
}

func (h *Handler) byIDs(c echo.Context) error {
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req ByIDsRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, err)
	}
	result, err := h.service.ByIDs(memberID, req)
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, result)
}

func queryInt(c echo.Context, key string) int {
	value := c.QueryParam(key)
	if value == "" {
		return 10
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 || parsed > 100 {
		return 10
	}
	return parsed
}
