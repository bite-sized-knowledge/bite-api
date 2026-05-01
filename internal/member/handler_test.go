package member

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
		"POST /v1/members",
		"POST /v1/members/join",
		"GET /v1/members/name/check",
		"DELETE /v1/members/:memberId",
	}

	for _, route := range expected {
		if !routes[route] {
			t.Fatalf("missing route %s", route)
		}
	}
}
