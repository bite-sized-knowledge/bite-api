package feed

import (
	"github.com/bite-sized/bite-api/internal/middleware"
	"github.com/bite-sized/bite-api/pkg/response"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func RegisterRoutes(v1 *echo.Group, h *Handler, authMiddleware echo.MiddlewareFunc, optionalAuth echo.MiddlewareFunc) {
	g := v1.Group("/feed")
	g.GET("", h.feed, optionalAuth)
}

// HeaderFeedRequestID — bite-web 가 응답 헤더에서 받아 sessionStorage 에 저장 →
// 다음 user_events 호출 시 feed_request_id 첨부 → impression ↔ click 정확 그룹핑.
const HeaderFeedRequestID = "X-Feed-Request-Id"

func (h *Handler) feed(c echo.Context) error {
	memberID, _ := middleware.CurrentMemberID(c)
	deviceID := c.Request().Header.Get(middleware.HeaderDeviceID)
	res, err := h.service.Feed(memberID, deviceID)
	if err != nil {
		return response.Error(c, err)
	}
	if res.FeedRequestID != "" {
		c.Response().Header().Set(HeaderFeedRequestID, res.FeedRequestID)
	}
	return response.Success(c, res.Items)
}
