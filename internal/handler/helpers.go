package handler

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/labstack/echo/v5"
	"github.com/starfederation/datastar-go/datastar"
)

var validate = validator.New()

// sse returns a Datastar SSE generator wired through Echo's raw response writer.
// In v5, c.Response() returns the http.ResponseWriter directly (no .Writer field).
func sse(c *echo.Context) *datastar.ServerSentEventGenerator {
	return datastar.NewSSE(c.Response(), c.Request())
}

// readSignals reads Datastar signals from the request body into target.
func readSignals(c *echo.Context, target any) error {
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

// formFieldName resolves the frontend signal name for a Go struct field, used to
// build the "<field>_error" signal key patched back on validation failure.
//
// It prefers the `json` tag (JSON/signals submission mode), then the `form` tag
// (multipart mode), then falls back to the lowercased Go field name. Tag options
// like `title,omitempty` are stripped so only the name remains.
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
	if name := cleanTagName(f.Tag.Get("json")); name != "" {
		return name
	}
	if name := cleanTagName(f.Tag.Get("form")); name != "" {
		return name
	}
	return strings.ToLower(goField)
}

// cleanTagName returns the tag name with any `,option` suffix removed, or "" for
// "-" / empty tags.
func cleanTagName(tag string) string {
	if tag == "" || tag == "-" {
		return ""
	}
	if i := strings.IndexByte(tag, ','); i >= 0 {
		tag = tag[:i]
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

func getSharedData(c *echo.Context) *domain.AdminSharedData {
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

// --- Action recipe helpers --------------------------------------------------
//
// The standard mutation handler recipe is:
//
//	decode → validate → map → service → domain-error map or toast → navigate/toast
//
// These wrappers each return error so handlers read as one line of intent,
// e.g. `return toast(c, "Saved", "success")` or
// `return patchFieldErrors(c, err, raw)`. They are thin wrappers over the raw
// sse() primitives above — no behavior change, just less boilerplate.

// toast emits a toast notification over SSE.
func toast(c *echo.Context, message, variant string) error {
	sse(c).ExecuteScript(toastScript(message, variant))
	return nil
}

// patchFieldErrors converts a validator error into per-field "<field>_error"
// signals and patches them over SSE.
func patchFieldErrors(c *echo.Context, err error, form any) error {
	sse(c).MarshalAndPatchSignals(fieldValidationErrors(err, form))
	return nil
}

// patchSignals patches arbitrary signal key/values over SSE.
func patchSignals(c *echo.Context, signals map[string]string) error {
	sse(c).MarshalAndPatchSignals(signals)
	return nil
}

// navigate runs window.navigateAdmin(url) over SSE, optionally followed by a
// toast. Pass message="" to navigate silently.
func navigate(c *echo.Context, url, message, variant string) error {
	script := fmt.Sprintf("window.navigateAdmin(%q)", url)
	if message != "" {
		script += "; " + toastScript(message, variant)
	}
	sse(c).ExecuteScript(script)
	return nil
}

// bustPagesCache prunes Swup's cached /admin/pages entry (and the specific
// editUrl when non-empty) so post-mutation navigation shows fresh content.
// An empty url is safe — the JS treats a falsy editUrl as list-only.
func bustPagesCache(c *echo.Context, url string) error {
	sse(c).ExecuteScript(fmt.Sprintf("window.bustPagesCache(%q)", url))
	return nil
}

// bustCache prunes a single URL from Swup's cache. Use after an in-place SSE
// mutation (no navigation) that changes a page's content, so back-navigation
// shows the fresh state rather than the pre-mutation snapshot.
func bustCache(c *echo.Context, url string) error {
	sse(c).ExecuteScript(fmt.Sprintf("window.bustCache(%q)", url))
	return nil
}
