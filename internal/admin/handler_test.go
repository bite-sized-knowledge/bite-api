package admin

import (
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRegisterRoutes(t *testing.T) {
	e := echo.New()
	g := e.Group("/admin")
	RegisterRoutes(g, &Handler{})

	routes := map[string]bool{}
	for _, route := range e.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	expected := []string{
		"DELETE /admin/articles/:articleId",
		"DELETE /admin/blogs/:blogId",
	}

	for _, route := range expected {
		if !routes[route] {
			t.Fatalf("missing route %s", route)
		}
	}
}
