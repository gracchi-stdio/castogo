package domain

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrDuplicateSlug = errors.New("episode slug already exists")
	ErrInvalidInput  = errors.New("invalid input")
)
