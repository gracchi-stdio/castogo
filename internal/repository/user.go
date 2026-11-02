package repository

import (
	"context"

	"github.com/gracchi-stdio/castogo/internal/domain"
)

// UserRepository is the storage contract for admin accounts. IDs are
// [16]byte — uuid.UUID is just a [16]byte type alias, so using the raw
// type avoids importing the uuid package in the interface.
type UserRepository interface {
	GetByID(ctx context.Context, id [16]byte) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) (*domain.User, error)
	Delete(ctx context.Context, id [16]byte) error
}
