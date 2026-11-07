package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/sqlcsqlite"
)

type UserRepo struct {
	q *sqlcsqlite.Queries
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{q: sqlcsqlite.New(db)}
}

func (ur UserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	row, err := ur.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return userFromRow(row)
}

func (ur UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := ur.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return userFromRow(row)
}

func (ur UserRepo) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	row, err := ur.q.CreateUser(ctx, sqlcsqlite.CreateUserParams{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	})
	if err != nil {
		return nil, err
	}
	return userFromRow(row)
}

func (ur UserRepo) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	row, err := ur.q.UpdateUser(ctx, sqlcsqlite.UpdateUserParams{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		ID:           user.ID,
	})
	if err != nil {
		return nil, err
	}
	return userFromRow(row)
}

func (ur UserRepo) Delete(ctx context.Context, id int64) error {
	return ur.q.DeleteUser(ctx, id)
}

func userFromRow(u sqlcsqlite.User) (*domain.User, error) {
	created, err := timeFromSQLite(u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse user created_at: %w", err)
	}

	return &domain.User{
		ID:           u.ID,
		PasswordHash: u.PasswordHash,
		Email:        u.Email,
		CreatedAt:    created,
		UpdatedAt:    sqliteTime(u.UpdatedAt),
	}, nil
}
