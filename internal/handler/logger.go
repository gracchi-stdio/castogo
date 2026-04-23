package handler

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
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

func RequestLogger(skipper func(echo.Context) bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			if skipper != nil && skipper(c) {
				return err
			}

			status := c.Response().Status
			if status == 0 {
				status = 200 // SSE via raw writer bypasses Echo's status tracking
			}

			method := c.Request().Method
			path := c.Request().URL.Path
			ms := time.Since(start).Milliseconds()

			fmt.Printf("%s%s%s %s%-6s%s %-20s %s%3d%s %s%4dms%s\n",
				cDim, start.Format("15:04:05"), cReset,
				cBold, method, cReset,
				path,
				statusColor(status), status, cReset,
				cDim, ms, cReset,
			)

			return err
		}
	}
}
