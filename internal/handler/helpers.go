package handler

import (
	"encoding/json"
	"reflect"
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

func fieldValidationErrors(err error, form any) map[string]string {
	result := map[string]string{}
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, e := range ve {
			field := formFieldName(form, e.Field())
			result[field+"_error"] = validationMsg(prettyFieldName(field), e.Tag(), e.Param())
		}
	} else {
		result["toast"] = err.Error()
	}
	return result
}

// formFieldName resolves the `form` tag for a Go struct field, falling back to
// the lowercased field name when the tag is missing. Used to align validation
// error signal keys with frontend signal names (snake_case).
func formFieldName(form any, goField string) string {
	t := reflect.TypeOf(form)
	if t == nil {
		return strings.ToLower(goField)
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	f, ok := t.FieldByName(goField)
	if !ok {
		return strings.ToLower(goField)
	}
	tag := f.Tag.Get("form")
	if tag == "" || tag == "-" {
		return strings.ToLower(goField)
	}
	return tag
}

// prettyFieldName converts a snake_case form field to space-separated words
// for use in user-facing messages.
func prettyFieldName(field string) string {
	return strings.ReplaceAll(field, "_", " ")
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

type toastPayload struct {
	Message string `json:"message"`
	Variant string `json:"variant"`
}

func toastScript(message, variant string) string {
	payload := toastPayload{Message: message, Variant: variant}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return "window.pushToast(" + string(encoded) + ")"
}
