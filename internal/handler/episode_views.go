package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/repository"
	episodeForm "github.com/gracchi-stdio/castogo/internal/view/editors/episode"
	"github.com/gracchi-stdio/castogo/internal/view/episodeview"
	"github.com/labstack/echo/v5"
)

// episodesList renders the admin episode list.
func (h *AdminHandler) episodesList(c *echo.Context) error {
	searchString := c.QueryParam("filter")
	offset := 0
	if offsetParam := c.QueryParam("offset"); offsetParam != "" {
		offset = parseInt(offsetParam)
	}

	episodes, err := h.episodeService.List(c.Request().Context(), repository.EpisodeFilter{
		Search: searchString,
		Limit:  100,
		Offset: offset,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load episodes")
	}
	return echo.WrapHandler(templ.Handler(episodeview.EpisodesListPage(getSharedData(c), episodes)))(c)
}

// episodeCreatePage renders the new-episode form.
func (h *AdminHandler) episodeCreatePage(c *echo.Context) error {
	return echo.WrapHandler(templ.Handler(episodeForm.Create(getSharedData(c))))(c)
}

// episodeEdit renders the episode edit page (metadata form + companion-page card).
func (h *AdminHandler) episodeEdit(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid episode ID")
	}

	episode, err := h.episodeService.GetByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Episode not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load episode")
	}

	pages, err := h.pageService.ListPages(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load pages")
	}

	args := episodeForm.Args{Episode: episode, Pages: pages}
	if episode.LinkedPageID != nil {
		for _, p := range pages {
			if p.ID == *episode.LinkedPageID {
				args.LinkedPage = p
				break
			}
		}
	}

	return echo.WrapHandler(templ.Handler(episodeForm.Edit(getSharedData(c), args)))(c)
}
