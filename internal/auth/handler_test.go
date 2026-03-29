package auth

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
		"POST /v1/auth/login",
		"POST /v1/auth/refresh",
		"POST /v1/auth/password/reset",
		"GET /v1/auth/email/verify",
		"POST /v1/auth/email/request-verify",
		"GET /v1/auth/email/is-verified",
		"POST /v1/auth/password/change",
		"POST /v1/auth/password/match",
	}

	for _, route := range expected {
		if !routes[route] {
			t.Fatalf("missing route %s", route)
		}
	}
}
