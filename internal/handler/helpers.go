package handler

import (
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"
)

var validate = validator.New()

// sse returns a Datastar SSE generator wired through Echo's raw response writer.
func sse(c echo.Context) *datastar.ServerSentEventGenerator {
	return datastar.NewSSE(c.Response().Writer, c.Request())
}

// readSignals reads Datastar signals from the request body into target.
func readSignals(c echo.Context, target any) error {
	return datastar.ReadSignals(c.Request(), target)
}

func fieldValidationErrors(err error) map[string]string {
	result := map[string]string{}
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, e := range ve {
			key := strings.ToLower(e.Field()) + "_error"
			result[key] = validationMsg(e.Field(), e.Tag(), e.Param())
		}
	} else {
		result["error"] = err.Error()
	}
	return result
}

func validationMsg(field, tag, param string) string {
	switch tag {
	case "required":
		return field + " is required"
	case "email":
		return "Please enter a valid " + strings.ToLower(field)
	case "min":
		return field + " must be at least " + param + " characters"
	case "url":
		return "Please enter a valid URL for " + strings.ToLower(field)
	default:
		return field + " is invalid"
	}
}

func getSharedData(c echo.Context) *domain.AdminSharedData {
	user, ok := c.Get("user").(*domain.User)
	if !ok || user == nil {
		return nil
	}

	return &domain.AdminSharedData{
		User:        user,
		CurrentPath: c.Request().URL.Path,
	}
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func parseInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}
