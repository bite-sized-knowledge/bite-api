package middleware

import (
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func CORS() echo.MiddlewareFunc {
	origins := []string{
		"https://bite-sized.xyz",
		"https://www.bite-sized.xyz",
		"https://dev.bite-sized.xyz",
	}
	if env := os.Getenv("APP_ENV"); env == "local" || env == "" {
		origins = append(origins, "http://localhost:3000", "http://localhost:3001")
	}
	if extra := os.Getenv("CORS_ALLOWED_ORIGINS"); extra != "" {
		origins = append(origins, strings.Split(extra, ",")...)
	}

	return echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     origins,
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
		MaxAge:           3600,
	})
}
