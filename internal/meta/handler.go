package meta

import (
	"github.com/bite-sized/bite-api/pkg/response"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func RegisterRoutes(v1 *echo.Group, h *Handler) {
	g := v1.Group("/meta")
	g.GET("/interests", h.interests)
}

func (h *Handler) interests(c echo.Context) error {
	rows, err := h.repo.ListInterests()
	if err != nil {
		return response.Error(c, err)
	}
	result := make([]InterestResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, InterestResponse{
			ID:        row.ID,
			Name:      row.Name,
			Image:     row.Image,
			Thumbnail: row.Thumbnail,
		})
	}
	return response.Success(c, result)
}
