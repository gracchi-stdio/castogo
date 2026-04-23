package handler

import (
	"time"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/gracchi-stdio/castogo/internal/view"
)

type ClockHandler struct{}

func NewClockHandler() *ClockHandler {
	return &ClockHandler{}
}

func (h *ClockHandler) RegisterRoutes(e *echo.Echo) {
	e.GET("/clock", echo.WrapHandler(templ.Handler(view.TiktakPage())))
	e.GET("/clock/stream", h.stream)
}

func (h *ClockHandler) stream(c echo.Context) error {
	out := sse(c)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now().Format("15:04:05")
			out.PatchElements(`<div id="clock">` + now + `</div>`)
		case <-c.Request().Context().Done():
			return nil
		}
	}
}
