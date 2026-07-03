// Package handler implements the HTTP handlers and route registration for the
//
// Handlers are built on Echo v4 and cover:
//   - Admin dashboard, settings, and CMS (pages, episodes)
//   - Authentication and session-based authorization
//   - Public page rendering and RSS feed generation
//
// Shared helpers (sse, readSignals, validate, fieldValidationErrors) provide
// Datastar SSE integration and request validation used across all handlers.
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/gracchi-stdio/castogo/internal/domain"
	pageform "github.com/gracchi-stdio/castogo/internal/view/editors/page"
	"github.com/labstack/echo/v4"
)

// create page view
func (h *AdminHandler) pageCreate(c echo.Context) error {
	parentPages, err := h.pageService.ListPages(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load parent pages")
	}

	return echo.WrapHandler(templ.Handler(pageform.Create(getSharedData(c), pageform.Args{
		ParentPages: parentPages,
	})))(c)
}

// page list view
func (h *AdminHandler) pageList(c echo.Context) error {
	pages, err := h.pageService.ListPages(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load pages")
	}
	return echo.WrapHandler(templ.Handler(pageform.List(getSharedData(c), pages)))(c)
}

// page edit view
func (h *AdminHandler) pageEdit(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page ID")
	}

	pageWithBlocks, err := h.pageService.GetPageWithBlocks(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Page not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load page")
	}

	parentPages, err := h.pageService.ListPages(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load parent pages")
	}

	return echo.WrapHandler(templ.Handler(pageform.EditSettings(
		getSharedData(c),
		pageform.Args{
			Page:        pageWithBlocks.Page,
			ParentPages: parentPages,
		},
	)))(c)
}
