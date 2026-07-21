package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *PublicHandler) RSSFeed(c *echo.Context) error {
	feed, err := h.feedService.BuildFeed(c.Request().Context())
	if err != nil {
		return err
	}

	resp := c.Response()
	resp.Header().Set(echo.HeaderContentType, "application/rss+xml; charset=utf-8")
	resp.WriteHeader(http.StatusOK)

	// The 200 status is already committed, so a mid-stream Write error can't be
	// turned into a clean HTTP error — returning it would make the centralized
	// error handler attempt a second WriteHeader. Log and swallow instead.
	if err := feed.Write(resp); err != nil {
		log.Printf("rss feed write failed: %v", err)
	}
	return nil
}
