package blog

import (
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRegisterRoutes(t *testing.T) {
	e := echo.New()
	v1 := e.Group("/v1")
	passthrough := func(next echo.HandlerFunc) echo.HandlerFunc { return next }
	RegisterRoutes(v1, &Handler{}, passthrough, passthrough, passthrough)

	routes := map[string]bool{}
	for _, route := range e.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	expected := []string{
		"GET /v1/blogs",
		"GET /v1/blogs/:blogId",
		"POST /v1/blogs/:blogId/subscribe",
		"DELETE /v1/blogs/:blogId/subscribe",
		"GET /v1/blogs/:blogId/articles",
	}

	for _, route := range expected {
		if !routes[route] {
			t.Fatalf("missing route %s", route)
		}
	}
}
