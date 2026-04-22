package handler

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

func writeSSE(c fiber.Ctx, event, data string) {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Write([]byte(fmt.Sprintf("event: %s\ndata: %s\n\n", event, data)))
}

func writeSignals(c fiber.Ctx, signals map[string]string) {
	payload, _ := json.Marshal(signals)
	writeSSE(c, "datastar-patch-signals", "signals "+string(payload))
}

func writeError(c fiber.Ctx, errMsg string) {
	writeSignals(c, map[string]string{"error": errMsg})
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
