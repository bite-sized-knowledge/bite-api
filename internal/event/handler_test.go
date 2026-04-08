package event

import (
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRegisterRoutes(t *testing.T) {
	e := echo.New()
	v1 := e.Group("/v1")

	noop := func(next echo.HandlerFunc) echo.HandlerFunc { return next }
	RegisterRoutes(v1, &Handler{}, noop, noop)

	routes := map[string]bool{}
	for _, route := range e.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	if !routes["POST /v1/events"] {
		t.Fatalf("missing route POST /v1/events")
	}
	if !routes["POST /v1/events/merge"] {
		t.Fatalf("missing route POST /v1/events/merge")
	}
}
