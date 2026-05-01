package article

import (
	"strconv"
	"strings"

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

func RegisterRoutes(v1 *echo.Group, h *Handler, authMiddleware echo.MiddlewareFunc, optionalAuth echo.MiddlewareFunc, lazyGuest echo.MiddlewareFunc) {
	g := v1.Group("/articles")

	// Anonymous + IP-based rate limit: search hits a FULLTEXT-indexed scan,
	// so cap bursts to protect the DB from automated scraping.
	searchRL := middleware.RateLimit(5, 10)

	// Public endpoints (optional auth)
	g.GET("/recent", h.recent, optionalAuth)
	g.GET("/search", h.search, searchRL)
	g.GET("/suggest", h.suggest)

	// FK-action endpoints: lazyGuest mints a guest member on first interaction
	// from an anonymous client (X-Device-Id required).
	g.POST("/:articleId/likes", h.like, optionalAuth, lazyGuest)
	g.DELETE("/:articleId/likes", h.unlike, optionalAuth, lazyGuest)
	g.POST("/:articleId/uninterests", h.uninterest, optionalAuth, lazyGuest)
	g.POST("/:articleId/bookmarks", h.bookmark, optionalAuth, lazyGuest)
	g.DELETE("/:articleId/bookmarks", h.unbookmark, optionalAuth, lazyGuest)
	g.POST("/:articleId/shares", h.share, optionalAuth, lazyGuest)

	protected := g.Group("")
	protected.Use(authMiddleware)
	protected.GET("/likes", h.likes)
	protected.GET("/bookmarks", h.bookmarks)
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

func (h *Handler) likes(c echo.Context) error {
	memberID, err := middleware.CurrentMemberID(c)
	if err != nil {
		return response.Error(c, err)
	}
	result, err := h.service.Likes(memberID, queryInt(c, "limit"), c.QueryParam("from"))
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, result)
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
	lang := c.QueryParam("lang")
	if lang != "ko" && lang != "en" {
		lang = ""
	}
	var blogIDs []int64
	if raw := c.QueryParam("blogId"); raw != "" {
		for _, s := range strings.Split(raw, ",") {
			s = strings.TrimSpace(s)
			if id, err := strconv.ParseInt(s, 10, 64); err == nil && id > 0 {
				blogIDs = append(blogIDs, id)
			}
		}
		if len(blogIDs) > 50 {
			blogIDs = blogIDs[:50]
		}
	}
	result, err := h.service.Recent(memberID, queryInt(c, "limit"), c.QueryParam("from"), lang, blogIDs)
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
	memberID, _ := middleware.CurrentMemberID(c)
	opts := SearchOptions{
		MemberID: memberID,
		Lang:     normalizedLang(c.QueryParam("lang")),
		Mode:     c.QueryParam("mode"),
	}
	if v := queryInt64Ptr(c, "category_id"); v != nil {
		opts.CategoryID = v
	}
	if v := queryInt64Ptr(c, "blog_id"); v != nil {
		opts.BlogID = v
	}
	if v := queryFloat64Ptr(c, "published_after"); v != nil {
		opts.PublishedAfter = v
	}
	if v := queryFloat64Ptr(c, "published_before"); v != nil {
		opts.PublishedBefore = v
	}
	result, err := h.service.Search(c.QueryParam("query"), queryInt(c, "limit"), c.QueryParam("from"), opts)
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, result)
}

func normalizedLang(raw string) string {
	if raw == "ko" || raw == "en" {
		return raw
	}
	return ""
}

func queryInt64Ptr(c echo.Context, key string) *int64 {
	raw := c.QueryParam(key)
	if raw == "" {
		return nil
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil
	}
	return &parsed
}

func queryFloat64Ptr(c echo.Context, key string) *float64 {
	raw := c.QueryParam(key)
	if raw == "" {
		return nil
	}
	parsed, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil
	}
	return &parsed
}

func (h *Handler) suggest(c echo.Context) error {
	prefix := c.QueryParam("q")
	limit := queryInt(c, "limit")
	suggestions := h.service.Suggest(prefix, limit)
	return response.Success(c, map[string]any{"suggestions": suggestions})
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
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}
