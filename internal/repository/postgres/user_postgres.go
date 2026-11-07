package postgres

import (
	"context"

	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/sqlcpg"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	q *sqlcpg.Queries
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		q: sqlcpg.New(pool),
	}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	params := sqlcpg.CreateUserParams{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	}

	result, err := r.q.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainUser(&result), nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	result, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return toDomainUser(&result), nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	result, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return toDomainUser(&result), nil
}

func (r *UserRepo) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	params := sqlcpg.UpdateUserParams{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	}
	result, err := r.q.UpdateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainUser(&result), nil
}

func (r *UserRepo) Delete(ctx context.Context, id int64) error {
	return r.q.DeleteUser(ctx, id)
}

// --- Type mapping helper ---
func toDomainUser(u *sqlcpg.User) *domain.User {
	return &domain.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt.Time,
		UpdatedAt:    u.UpdatedAt.Time,
	}
}
