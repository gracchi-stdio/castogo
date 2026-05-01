package postgres

import (
	"context"

	"github.com/gracchi-stdio/castogo/internal/db"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	q *db.Queries
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		q: db.New(pool),
	}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	params := db.CreateUserParams{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	}

	result, err := r.q.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainUser(&result), nil

}

func (r *UserRepo) GetByID(ctx context.Context, id [16]byte) (*domain.User, error) {
	result, err := r.q.GetUserByID(ctx, uuidToPgtype(id))
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
	params := db.UpdateUserParams{
		ID:           uuidToPgtype(user.ID),
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	}
	result, err := r.q.UpdateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainUser(&result), nil
}

func (r *UserRepo) Delete(ctx context.Context, id [16]byte) error {
	return r.q.DeleteUser(ctx, uuidToPgtype(id))
}

// --- Type mapping helper ---
func toDomainUser(u *db.User) *domain.User {
	return &domain.User{
		ID:           pgtypeToUUID(u.ID),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
	}
}

func uuidToPgtype(uuid [16]byte) pgtype.UUID {
	return pgtype.UUID{
		Bytes: uuid,
		Valid: true,
	}
}

func pgtypeToUUID(u pgtype.UUID) [16]byte {
	return u.Bytes
}
