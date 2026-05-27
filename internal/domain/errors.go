package domain

import "errors"

var (
	// Shared
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidInput = errors.New("invalid input")

	// Episode
	ErrDuplicateSlug = errors.New("episode slug already exists")

	// Page
	ErrReservedSlug   = errors.New("slug is reserved")
	ErrDuplicatePath  = errors.New("page path already exists")
	ErrMaxDepth       = errors.New("maximum nesting depth exceeded")
	ErrInvalidParent  = errors.New("invalid parent page")
	ErrHomepageExists = errors.New("a homepage page already exists — only one root page may have an empty slug")
)
