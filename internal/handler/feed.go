package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *PunlicHandler) RSSFeed(c echo.Context) error {
	feed, err := h.feedService.BuildFeed(c.Request().Context())
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentType, "application/rss+xml; charset=utf-8")
	c.Response().WriteHeader(http.StatusOK)

	return feed.Write(c.Response().Writer)
}
