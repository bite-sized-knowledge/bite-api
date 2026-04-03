package response

import (
	"errors"
	"net/http"

	"github.com/bite-sized/bite-api/internal/model"
	"github.com/labstack/echo/v4"
)

type APIResponse struct {
	Success bool    `json:"success"`
	Result  unknown `json:"result,omitempty"`
}

type ApiError struct {
	Message string `json:"message"`
}

type unknown = interface{}

func Success(c echo.Context, data unknown) error {
	return c.JSON(http.StatusOK, APIResponse{Success: true, Result: data})
}

func Created(c echo.Context, data unknown) error {
	return c.JSON(http.StatusCreated, APIResponse{Success: true, Result: data})
}

func NoContent(c echo.Context) error {
	return c.JSON(http.StatusOK, APIResponse{Success: true})
}

func Error(c echo.Context, err error) error {
	status := http.StatusInternalServerError
	message := "internal server error"

	switch {
	case errors.Is(err, model.ErrBadRequest):
		status = http.StatusBadRequest
		message = err.Error()
	case errors.Is(err, model.ErrNotFound):
		status = http.StatusNotFound
		message = err.Error()
	case errors.Is(err, model.ErrConflict):
		status = http.StatusConflict
		message = err.Error()
	case errors.Is(err, model.ErrForbidden):
		status = http.StatusForbidden
		message = err.Error()
	case err != nil:
		// Do not expose internal error details
	}

	return c.JSON(status, APIResponse{Success: false, Result: &ApiError{Message: message}})
}
