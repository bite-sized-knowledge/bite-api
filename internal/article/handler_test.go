package article

import (
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRegisterRoutes(t *testing.T) {
	e := echo.New()
	v1 := e.Group("/v1")
	RegisterRoutes(v1, &Handler{})

	routes := map[string]bool{}
	for _, route := range e.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	expected := []string{
		"POST /v1/articles/:articleId/likes",
		"DELETE /v1/articles/:articleId/likes",
		"POST /v1/articles/:articleId/uninterests",
		"GET /v1/articles/bookmarks",
		"POST /v1/articles/:articleId/bookmarks",
		"DELETE /v1/articles/:articleId/bookmarks",
		"POST /v1/articles/:articleId/shares",
		"GET /v1/articles/recent",
		"GET /v1/articles/history",
		"GET /v1/articles/search",
		"POST /v1/articles/by-ids",
	}

	for _, route := range expected {
		if !routes[route] {
			t.Fatalf("missing route %s", route)
		}
	}
}
