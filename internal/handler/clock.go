package handler

import (
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gracchi-stdio/castogo/internal/view"
	"github.com/starfederation/datastar-go/datastar"
)

type ClockHandler struct{}

func NewClockHandler() *ClockHandler {
	return &ClockHandler{}
}

func (h *ClockHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/clock", templ.Handler(view.TiktakPage()))
	app.Get("/clock/stream", adaptor.HTTPHandlerFunc(h.streamHTTP))
}

func (h *ClockHandler) streamHTTP(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now().Format("15:04:05")
			sse.PatchElements(`<div id="clock">` + now + `</div>`)
		case <-r.Context().Done():
			return
		}
	}
}
