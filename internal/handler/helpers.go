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
	default:
		return field + " is invalid"
	}
}

func getSharedData(c echo.Context) *domain.AdminSharedData {
	user := c.Get("user").(*domain.User)
	return &domain.AdminSharedData{
		User: user,
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

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		if r == ' ' || r == '-' || r == '_' {
			return '-'
		}
		return -1
	}, s)
	return strings.Trim(s, "-")
}
