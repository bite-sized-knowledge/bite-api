package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

func RateLimit(r rate.Limit, b int) echo.MiddlewareFunc {
	return echomw.RateLimiter(echomw.NewRateLimiterMemoryStoreWithConfig(
		echomw.RateLimiterMemoryStoreConfig{Rate: r, Burst: b, ExpiresIn: 30 * time.Minute},
	))
}
