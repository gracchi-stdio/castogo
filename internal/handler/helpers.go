package handler

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

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
