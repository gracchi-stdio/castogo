package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/view/admin/page"
	"github.com/labstack/echo/v5"
)

// create page view
func (h *AdminHandler) pageCreate(c *echo.Context) error {
	parentPages, err := h.pageService.ListPages(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load parent pages")
	}

	return echo.WrapHandler(templ.Handler(page.Create(getSharedData(c), page.Args{
		ParentPages: parentPages,
	})))(c)
}

// page list view
func (h *AdminHandler) pageList(c *echo.Context) error {
	pages, err := h.pageService.ListPages(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load pages")
	}
	return echo.WrapHandler(templ.Handler(page.List(getSharedData(c), pages)))(c)
}

// page edit view
func (h *AdminHandler) pageEdit(c *echo.Context) error {
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

	return echo.WrapHandler(templ.Handler(page.EditSettings(
		getSharedData(c),
		page.Args{
			Page:        pageWithBlocks.Page,
			ParentPages: parentPages,
		},
	)))(c)
}

func (h *AdminHandler) blockEdit(c *echo.Context) error {
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

	return echo.WrapHandler(templ.Handler(page.EditBlocks(
		getSharedData(c),
		page.Args{
			Page:   pageWithBlocks.Page,
			Blocks: pageWithBlocks.Blocks,
		},
	)))(c)
}
