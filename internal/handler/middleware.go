package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gracchi-stdio/castogo/internal/repository"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func redirectLogin(c echo.Context) error {
	return c.Redirect(http.StatusFound, "/login")
}

func AuthMiddleware(userRepo repository.UserRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, _ := session.Get("session", c)

			idStr, _ := sess.Values["user_id"].(string)
			if idStr == "" {
				return redirectLogin(c)
			}

			id, err := uuid.Parse(idStr)
			if err != nil {
				return redirectLogin(c)
			}

			user, err := userRepo.GetByID(c.Request().Context(), id)
			if err != nil {
				return redirectLogin(c)
			}

			c.Set("user", user)
			return next(c)
		}
	}
}
