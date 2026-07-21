package handler

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v5"
)

const (
	cReset  = "\033[0m"
	cRed    = "\033[31m"
	cGreen  = "\033[32m"
	cYellow = "\033[33m"
	cCyan   = "\033[36m"
	cBold   = "\033[1m"
	cDim    = "\033[2m"
)

func statusColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return cGreen
	case code >= 300 && code < 400:
		return cCyan
	case code >= 400 && code < 500:
		return cYellow
	default:
		return cRed
	}
}

func RequestLogger(skipper func(*echo.Context) bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			err := next(c)

			if skipper != nil && skipper(c) {
				return err
			}

			// In v5, c.Response() returns the *echo.Response wrapper (not the raw
			// writer). Reach it via UnwrapResponse to read its Status field. SSE
			// handlers write through the raw writer and mark the wrapper Committed
			// without setting Status, so Status stays 0 — treat that as 200.
			var status int
			if resp, _ := echo.UnwrapResponse(c.Response()); resp != nil {
				status = resp.Status
			}
			if status == 0 {
				status = 200
			}

			method := c.Request().Method
			path := c.Request().URL.Path
			ms := time.Since(start).Milliseconds()

			// Short request id (set by middleware.RequestID, which runs before
			// this logger) for correlating a log line with downstream traces.
			reqID := c.Response().Header().Get(echo.HeaderXRequestID)
			if len(reqID) > 8 {
				reqID = reqID[:8]
			}

			fmt.Printf("%s%s%s %s%-8s%s %s%-6s%s %-20s %s%3d%s %s%4dms%s\n",
				cDim, start.Format("15:04:05"), cReset,
				cDim, reqID, cReset,
				cBold, method, cReset,
				path,
				statusColor(status), status, cReset,
				cDim, ms, cReset,
			)

			return err
		}
	}
}
