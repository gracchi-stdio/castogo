package repository

import (
	"context"

	"github.com/gracchi-stdio/podlog/internal/domain"
)

//	Note: [16]byte for UUID — uuid.UUID is just a [16]byte
//
// type alias. Using the raw type avoids importing the uuid
// package in the interface.
type UserRepository interface {
	GetByID(ctx context.Context, id [16]byte) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) (*domain.User, error)
	Delete(ctx context.Context, id [16]byte) error
}
