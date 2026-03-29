package middleware

import (
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func Recover() echo.MiddlewareFunc {
	return echomw.Recover()
}
