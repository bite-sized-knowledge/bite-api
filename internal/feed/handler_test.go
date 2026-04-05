package feed

import (
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRegisterRoutes(t *testing.T) {
	e := echo.New()
	v1 := e.Group("/v1")
	passthrough := func(next echo.HandlerFunc) echo.HandlerFunc { return next }
	RegisterRoutes(v1, &Handler{}, passthrough, passthrough)

	routes := map[string]bool{}
	for _, route := range e.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	if !routes["GET /v1/feed"] {
		t.Fatalf("missing route GET /v1/feed")
	}
}
